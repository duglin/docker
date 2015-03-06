package builder

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (b *Builder) ProcessDockerScript(filename string) (bool, error) {
	fmt.Printf("In ProcessDockerScript: %s\n", filename)
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	image := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#!") {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			return false, nil
		}
		line = strings.TrimSpace(line[1:])
		if line == "" { // Be very forgiving - let "#" pass
			continue
		}
		if !strings.HasPrefix(strings.ToUpper(line), "FROM:") || len(line) < 6 {
			return false, nil
		}
		image = strings.TrimSpace(line[5:])
		if i := strings.Index(image, "#"); i >= 0 {
			image = strings.TrimSpace(image[:i])
		}
		break
	}
	if image == "" {
		return false, nil
	}

	fmt.Fprintf(b.OutStream, "Running build from image:%s\n", image)

	b.UtilizeCache = false

	err = from(b, []string{image}, nil, "FROM "+image)
	if err != nil {
		return false, err
	}

	scriptName := filename[len(b.contextPath):]
	if scriptName[0] == '/' {
		scriptName = scriptName[1:]
	}

	args := []string{"cd /src && PATH=/src/.dbin:$PATH sh -e ./" + scriptName}

	err = run(b, args, nil, "RUN "+args[0])
	if err != nil {
		return true, err
	}

	// b.disableCommit = true

	return true, nil
}
