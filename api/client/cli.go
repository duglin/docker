package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/pkg/homedir"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
)

type ConfigFile struct {
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
	filename    string
}

type DockerCli struct {
	proto      string
	addr       string
	configFile *ConfigFile
	authFile   *registry.ConfigFile
	in         io.ReadCloser
	out        io.Writer
	err        io.Writer
	keyFile    string
	tlsConfig  *tls.Config
	scheme     string
	// inFd holds file descriptor of the client's STDIN, if it's a valid file
	inFd uintptr
	// outFd holds file descriptor of the client's STDOUT, if it's a valid file
	outFd uintptr
	// isTerminalIn describes if client's STDIN is a TTY
	isTerminalIn bool
	// isTerminalOut describes if client's STDOUT is a TTY
	isTerminalOut bool
	transport     *http.Transport
}

var funcMap = template.FuncMap{
	"json": func(v interface{}) string {
		a, _ := json.Marshal(v)
		return string(a)
	},
}

func (cli *DockerCli) getMethod(args ...string) (func(...string) error, bool) {
	camelArgs := make([]string, len(args))
	for i, s := range args {
		if len(s) == 0 {
			return nil, false
		}
		camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	methodName := "Cmd" + strings.Join(camelArgs, "")
	method := reflect.ValueOf(cli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

// Cmd executes the specified command.
func (cli *DockerCli) Cmd(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			return method(args[2:]...)
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Fprintf(cli.err, "docker: '%s' is not a docker command. See 'docker --help'.\n", args[0])
			os.Exit(1)
		}
		return method(args[1:]...)
	}
	return cli.CmdHelp()
}

func (cli *DockerCli) Subcmd(name, signature, description string, exitOnError bool) *flag.FlagSet {
	var errorHandling flag.ErrorHandling
	if exitOnError {
		errorHandling = flag.ExitOnError
	} else {
		errorHandling = flag.ContinueOnError
	}
	flags := flag.NewFlagSet(name, errorHandling)
	flags.Usage = func() {
		options := ""
		if signature != "" {
			signature = " " + signature
		}
		if flags.FlagCountUndeprecated() > 0 {
			options = " [OPTIONS]"
		}
		fmt.Fprintf(cli.out, "\nUsage: docker %s%s%s\n\n%s\n\n", name, options, signature, description)
		flags.SetOutput(cli.out)
		flags.PrintDefaults()
		os.Exit(0)
	}
	return flags
}

func (cli *DockerCli) LoadAuthFile() (err error) {
	cli.authFile, err = registry.LoadConfig(homedir.Get())
	if err != nil {
		fmt.Fprintf(cli.err, "WARNING: %s\n", err)
	}
	return err
}

func (cli *DockerCli) CheckTtyInput(attachStdin, ttyMode bool) error {
	// In order to attach to a container tty, input stream for the client must
	// be a tty itself: redirecting or piping the client standard input is
	// incompatible with `docker run -t`, `docker exec -t` or `docker attach`.
	if ttyMode && attachStdin && !cli.isTerminalIn {
		return errors.New("cannot enable tty mode on non tty input")
	}
	return nil
}

func NewDockerCli(in io.ReadCloser, out, err io.Writer, keyFile string, proto, addr string, tlsConfig *tls.Config) *DockerCli {
	var (
		inFd          uintptr
		outFd         uintptr
		isTerminalIn  = false
		isTerminalOut = false
		scheme        = "http"
	)

	if tlsConfig != nil {
		scheme = "https"
	}
	if in != nil {
		inFd, isTerminalIn = term.GetFdInfo(in)
	}

	if out != nil {
		outFd, isTerminalOut = term.GetFdInfo(out)
	}

	if err == nil {
		err = out
	}

	// The transport is created here for reuse during the client session
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Why 32? See issue 8035
	timeout := 32 * time.Second
	if proto == "unix" {
		// no need in compressing for local communications
		tr.DisableCompression = true
		tr.Dial = func(_, _ string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	} else {
		tr.Proxy = http.ProxyFromEnvironment
		tr.Dial = (&net.Dialer{Timeout: timeout}).Dial
	}

	configFile, e := LoadConfigFile(filepath.Join(homedir.Get(), ".docker"))
	if e != nil {
		fmt.Fprintf(err, "WARNING: Error loading config file:%s\n", e)
	}

	return &DockerCli{
		proto:         proto,
		addr:          addr,
		configFile:    configFile,
		in:            in,
		out:           out,
		err:           err,
		keyFile:       keyFile,
		inFd:          inFd,
		outFd:         outFd,
		isTerminalIn:  isTerminalIn,
		isTerminalOut: isTerminalOut,
		tlsConfig:     tlsConfig,
		scheme:        scheme,
		transport:     tr,
	}
}

func LoadConfigFile(dirPath string) (*ConfigFile, error) {
	fn := filepath.Join(dirPath, "config.json")

	configFile := ConfigFile{
		filename: fn,
	}

	if _, err := os.Stat(fn); err != nil {
		// no file is not an error
		return &configFile, nil
	}

	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return &configFile, fmt.Errorf("Error reading config file(%s):%s", fn, err)
	}

	err = json.Unmarshal(data, &configFile)
	if err != nil {
		return &configFile, fmt.Errorf("Error parsing config file(%s): %s", fn, err)
	}

	return &configFile, nil
}

func (cli *DockerCli) SaveConfigFile() error {
	data, err := json.MarshalIndent(cli.configFile, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cli.configFile.filename, data, 0600)
	if err != nil {
		return err
	}
	return nil
}
