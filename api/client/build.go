package client

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/jsonmessage"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/pkg/progressreader"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/pkg/symlink"
	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
)

const (
	tarHeaderSize = 512
)

// CmdBuild builds a new image from the source code at a given path.
//
// If '-' is provided instead of a path or URL, Docker will build an image from either a Dockerfile or tar archive read from STDIN.
//
// Usage: docker build [OPTIONS] PATH | URL | -
func (cli *DockerCli) CmdBuild(args ...string) error {
	cmd := cli.Subcmd("build", "PATH | URL | -", "Build a new image from the source code at PATH", true)
	tag := cmd.String([]string{"t", "-tag"}, "", "Repository name (and optionally a tag) for the image")
	suppressOutput := cmd.Bool([]string{"q", "-quiet"}, false, "Suppress the verbose output generated by the containers")
	noCache := cmd.Bool([]string{"#no-cache", "-no-cache"}, false, "Do not use cache when building the image")
	rm := cmd.Bool([]string{"#rm", "-rm"}, true, "Remove intermediate containers after a successful build")
	forceRm := cmd.Bool([]string{"-force-rm"}, false, "Always remove intermediate containers")
	pull := cmd.Bool([]string{"-pull"}, false, "Always attempt to pull a newer version of the image")
	dockerfileName := cmd.String([]string{"f", "-file"}, "", "Name of the Dockerfile (Default is 'PATH/Dockerfile')")
	flMemoryString := cmd.String([]string{"m", "-memory"}, "", "Memory limit")
	flMemorySwap := cmd.String([]string{"-memory-swap"}, "", "Total memory (memory + swap), '-1' to disable swap")
	flCPUShares := cmd.Int64([]string{"c", "-cpu-shares"}, 0, "CPU shares (relative weight)")
	flCPUSetCpus := cmd.String([]string{"-cpuset-cpus"}, "", "CPUs in which to allow execution (0-3, 0,1)")

	cmd.Require(flag.Exact, 1)

	utils.ParseFlags(cmd, args, true)

	var (
		context  archive.Archive
		isRemote bool
		err      error
	)

	_, err = exec.LookPath("git")
	hasGit := err == nil
	if cmd.Arg(0) == "-" {
		// As a special case, 'docker build -' will build from either an empty context with the
		// contents of stdin as a Dockerfile, or a tar-ed context from stdin.
		buf := bufio.NewReader(cli.in)
		magic, err := buf.Peek(tarHeaderSize)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to peek context header from STDIN: %v", err)
		}
		if !archive.IsArchive(magic) {
			dockerfile, err := ioutil.ReadAll(buf)
			if err != nil {
				return fmt.Errorf("failed to read Dockerfile from STDIN: %v", err)
			}

			// -f option has no meaning when we're reading it from stdin,
			// so just use our default Dockerfile name
			*dockerfileName = api.DefaultDockerfileName
			context, err = archive.Generate(*dockerfileName, string(dockerfile))
		} else {
			context = ioutil.NopCloser(buf)
		}
	} else if urlutil.IsURL(cmd.Arg(0)) && (!urlutil.IsGitURL(cmd.Arg(0)) || !hasGit) {
		isRemote = true
	} else {
		root := cmd.Arg(0)
		if urlutil.IsGitURL(root) {
			remoteURL := cmd.Arg(0)
			if !urlutil.IsGitTransport(remoteURL) {
				remoteURL = "https://" + remoteURL
			}

			root, err = ioutil.TempDir("", "docker-build-git")
			if err != nil {
				return err
			}
			defer os.RemoveAll(root)

			if output, err := exec.Command("git", "clone", "--recursive", remoteURL, root).CombinedOutput(); err != nil {
				return fmt.Errorf("Error trying to use git: %s (%s)", err, output)
			}
		}
		if _, err := os.Stat(root); err != nil {
			return err
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		filename := *dockerfileName // path to Dockerfile

		if *dockerfileName == "" {
			// No -f/--file was specified so use the default
			*dockerfileName = api.DefaultDockerfileName
			filename = filepath.Join(absRoot, *dockerfileName)

			// Just to be nice ;-) look for 'dockerfile' too but only
			// use it if we found it, otherwise ignore this check
			if _, err = os.Lstat(filename); os.IsNotExist(err) {
				tmpFN := path.Join(absRoot, strings.ToLower(*dockerfileName))
				if _, err = os.Lstat(tmpFN); err == nil {
					*dockerfileName = strings.ToLower(*dockerfileName)
					filename = tmpFN
				}
			}
		}

		origDockerfile := *dockerfileName // used for error msg
		if filename, err = filepath.Abs(filename); err != nil {
			return err
		}

		// Verify that 'filename' is within the build context
		filename, err = symlink.FollowSymlinkInScope(filename, absRoot)
		if err != nil {
			return fmt.Errorf("The Dockerfile (%s) must be within the build context (%s)", origDockerfile, root)
		}

		// Now reset the dockerfileName to be relative to the build context
		*dockerfileName, err = filepath.Rel(absRoot, filename)
		if err != nil {
			return err
		}
		// And canonicalize dockerfile name to a platform-independent one
		*dockerfileName, err = archive.CanonicalTarNameForPath(*dockerfileName)
		if err != nil {
			return fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
		}

		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			return fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
		}
		var includes = []string{"."}

		excludes, err := utils.ReadDockerIgnore(path.Join(root, ".dockerignore"))
		if err != nil {
			return err
		}

		// If .dockerignore mentions .dockerignore or the Dockerfile
		// then make sure we send both files over to the daemon
		// because Dockerfile is, obviously, needed no matter what, and
		// .dockerignore is needed to know if either one needs to be
		// removed.  The deamon will remove them for us, if needed, after it
		// parses the Dockerfile.
		keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
		keepThem2, _ := fileutils.Matches(*dockerfileName, excludes)
		if keepThem1 || keepThem2 {
			includes = append(includes, ".dockerignore", *dockerfileName)
		}

		if err = utils.ValidateContextDirectory(root, excludes); err != nil {
			return fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
		}
		options := &archive.TarOptions{
			Compression:     archive.Uncompressed,
			ExcludePatterns: excludes,
			IncludeFiles:    includes,
		}
		context, err = archive.TarWithOptions(root, options)
		if err != nil {
			return err
		}
	}

	// windows: show error message about modified file permissions
	// FIXME: this is not a valid warning when the daemon is running windows. should be removed once docker engine for windows can build.
	if runtime.GOOS == "windows" {
		log.Warn(`SECURITY WARNING: You are building a Docker image from Windows against a Linux Docker host. All files and directories added to build context will have '-rwxr-xr-x' permissions. It is recommended to double check and reset permissions for sensitive files and directories.`)
	}

	var body io.Reader
	// Setup an upload progress bar
	// FIXME: ProgressReader shouldn't be this annoying to use
	if context != nil {
		sf := streamformatter.NewStreamFormatter(false)
		body = progressreader.New(progressreader.Config{
			In:        context,
			Out:       cli.out,
			Formatter: sf,
			NewLines:  true,
			ID:        "",
			Action:    "Sending build context to Docker daemon",
		})
	}

	var memory int64
	if *flMemoryString != "" {
		parsedMemory, err := units.RAMInBytes(*flMemoryString)
		if err != nil {
			return err
		}
		memory = parsedMemory
	}

	var memorySwap int64
	if *flMemorySwap != "" {
		if *flMemorySwap == "-1" {
			memorySwap = -1
		} else {
			parsedMemorySwap, err := units.RAMInBytes(*flMemorySwap)
			if err != nil {
				return err
			}
			memorySwap = parsedMemorySwap
		}
	}
	// Send the build context
	v := &url.Values{}

	//Check if the given image name can be resolved
	if *tag != "" {
		repository, tag := parsers.ParseRepositoryTag(*tag)
		if err := registry.ValidateRepositoryName(repository); err != nil {
			return err
		}
		if len(tag) > 0 {
			if err := graph.ValidateTagName(tag); err != nil {
				return err
			}
		}
	}

	v.Set("t", *tag)

	if *suppressOutput {
		v.Set("q", "1")
	}
	if isRemote {
		v.Set("remote", cmd.Arg(0))
	}
	if *noCache {
		v.Set("nocache", "1")
	}
	if *rm {
		v.Set("rm", "1")
	} else {
		v.Set("rm", "0")
	}

	if *forceRm {
		v.Set("forcerm", "1")
	}

	if *pull {
		v.Set("pull", "1")
	}

	v.Set("cpusetcpus", *flCPUSetCpus)
	v.Set("cpushares", strconv.FormatInt(*flCPUShares, 10))
	v.Set("memory", strconv.FormatInt(memory, 10))
	v.Set("memswap", strconv.FormatInt(memorySwap, 10))

	v.Set("dockerfile", *dockerfileName)

	headers := http.Header(make(map[string][]string))
	buf, err := json.Marshal(cli.configFile)
	if err != nil {
		return err
	}
	headers.Add("X-Registry-Config", base64.URLEncoding.EncodeToString(buf))

	if context != nil {
		headers.Set("Content-Type", "application/tar")
	}
	err = cli.stream("POST", fmt.Sprintf("/build?%s", v.Encode()), body, cli.out, headers)
	if jerr, ok := err.(*jsonmessage.JSONError); ok {
		// If no error code is set, default to 1
		if jerr.Code == 0 {
			jerr.Code = 1
		}
		return &utils.StatusError{Status: jerr.Message, StatusCode: jerr.Code}
	}
	return err
}
