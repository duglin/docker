package envutil

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestEnvProcessor(t *testing.T) {
	file, err := os.Open("words")
	if err != nil {
		t.Fatalf("Can't open 'words': %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	envs := []string{"PWD=/home", "SHELL=bash"}
	for scanner.Scan() {
		line := scanner.Text()

		// Trim comments and blank lines
		i := strings.Index(line, "#")
		if i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		words := strings.Split(line, "|")
		if len(words) != 2 {
			t.Fatalf("Error in 'words' - should be 2 words:%q", words)
		}

		words[0] = strings.TrimSpace(words[0])
		words[1] = strings.TrimSpace(words[1])

		newWord, err := ProcessWord(words[0], envs)

		if err != nil {
			newWord = "error"
		}

		if newWord != words[1] {
			t.Fatalf("Error. Src: %s  Calc: %s  Expected: %s", words[0], newWord, words[1])
		}
	}
}

func TestReplaceAndAppendEnvVars(t *testing.T) {
	var (
		d = []string{"HOME=/"}
		o = []string{"HOME=/root", "TERM=xterm"}
	)

	env := ReplaceOrAppendEnvValues(d, o)
	if len(env) != 2 {
		t.Fatalf("expected len of 2 got %d", len(env))
	}
	if env[0] != "HOME=/root" {
		t.Fatalf("expected HOME=/root got '%s'", env[0])
	}
	if env[1] != "TERM=xterm" {
		t.Fatalf("expected TERM=xterm got '%s'", env[1])
	}
}
