package jsondiff

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	// For each *.test file extract the 2 JSON files and the expected results

	files, err := ioutil.ReadDir("tests")
	if err != nil {
		t.Fatalf("Error loading dir: %s", err)
	}

	for _, fileInfo := range files {
		fn := fileInfo.Name()
		if fileInfo.IsDir() || len(fn) < 5 || fn[len(fn)-4:] != ".tst" {
			continue
		}

		fmt.Printf("  Testing: %s\n", fn)

		fn = "tests/" + fn
		file, err := os.Open(fn)
		if err != nil {
			t.Fatalf("Error opening %s: %s", fn, err)
		}

		data := []string{"", "", "", ""}

		phase := 0 // 0 = result, 1 == options, 2 == file1, 3 == file2
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if i := strings.Index(line, "// "); i >= 0 {
				line = line[:i]
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if line == "===" {
				phase++
				continue
			}
			data[phase] += line + "\n"
		}

		f1, err := ioutil.TempFile("", "json1.")
		if err != nil {
			t.Fatalf("Error creating temp file: %s", err)
		}
		defer os.Remove(f1.Name())
		f1.WriteString(data[2])
		f1.Close()

		f2, err := ioutil.TempFile("", "json2.")
		if err != nil {
			t.Fatalf("Error creating temp file: %s", err)
		}
		defer os.Remove(f2.Name())
		f2.WriteString(data[3])
		f2.Close()

		diff, err := diffFiles(f1.Name(), f2.Name())
		if diff != data[0] || (err != nil && err.Error() != data[0]) {
			t.Fatalf("\nUnexpected failure on %q:\nExpected:%sGot:\nErr:%s\n\n%s", fn, data[0], err, diff)
		}
	}
}
