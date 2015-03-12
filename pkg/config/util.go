package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type Context struct {
	dir string
}

func NewContext(files map[string]string) (*Context, error) {
	dir, err := ioutil.TempDir("", "cfg")
	if err != nil {
		return nil, err
	}
	ctx := &Context{dir: dir}
	if files != nil {
		for file := range files {
			path := filepath.Join(ctx.dir, file)
			data := []byte(files[file])
			err := ioutil.WriteFile(path, []byte(data), 0700)
			if err != nil {
				ctx.Delete()
				return nil, err
			}
		}
	}
	return ctx, nil
}

func (ctx *Context) Delete() {
	if ctx.dir != "" {
		os.RemoveAll(ctx.dir)
	}
}

func (ctx *Context) Dir() string {
	return ctx.dir
}

func (ctx *Context) File(fn string) string {
	return filepath.Join(ctx.dir, fn)
}

// 'expected and 'got' MUST match - exactly
func CheckAllMap(got, expected map[string]string) error {
	if len(expected) != len(got) {
		return fmt.Errorf("Diff in # of fields\nExpected:%q\nGot:%q",
			expected, got)
	}
	for key := range expected {
		if got[key] != expected[key] {
			return fmt.Errorf("Unexcepted data for %q: %q - should be %q",
				key, got[key], expected[key])
		}
	}
	return nil
}

// Just check the feilds specified in 'got'
func CheckSomeMap(got, expected map[string]string) error {
	for key := range got {
		if got[key] != expected[key] {
			return fmt.Errorf("Unexcepted data for %q: %q - should be %q",
				key, got[key], expected[key])
		}
	}
	return nil
}

func CheckSlice(oldGot, expected []string) error {
	if len(oldGot) != len(expected) {
		return fmt.Errorf("Diff in size of slices:\nGot(%d):%s\nExpected(%d):%s", len(oldGot), oldGot, len(expected), expected)
	}

	got := []string{}
	for _, val := range oldGot {
		got = append(got, val)
	}

	for _, val := range expected {
		found := false
		for i, val2 := range got {
			found = (val2 == val)
			if found {
				got = append(got[:i], got[i+1:]...)
				break
			}
		}
		if !found {
			return fmt.Errorf("Missing %q in slice", val)
		}
	}
	return nil
}

func pass() {
	pc := make([]uintptr, 10)
	runtime.Callers(0, pc)
	name := runtime.FuncForPC(pc[2]).Name()
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	fmt.Printf("  PASS - config: %s\n", name)
}

func Check(t *testing.T, cfg *Config, field string, cfgField interface{}, val interface{}) {
	var errStr string

	for {
		if fmt.Sprintf("%T", cfgField) != fmt.Sprintf("%T", val) {
			errStr = fmt.Sprintf("Types of cfgField and val are not the same!!")
			break
		}

		if cfgField != val {
			errStr = fmt.Sprintf("Data field(%s) value(%v) != expected value: %q", field, cfgField, val)
			break
		}

		v, err := cfg.Get(field)
		if err != nil {
			errStr = err.Error()
			break
		}

		if v != fmt.Sprintf("%v", val) {
			errStr = fmt.Sprintf("Get(%s) != expected value: %v", field, val)
			break
		}
		break
	}

	if errStr == "" {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(0, pc)
	fn := runtime.FuncForPC(pc[2])
	file, line := fn.FileLine(pc[2])

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	t.Fatalf("\n*** Error: %s\n*** From before: %s:%d\n",
		errStr, file, line)
}

func Set(t *testing.T, cfg *Config, key string, val interface{}) {
	str := fmt.Sprintf("%v", val)

	err := cfg.Set(key, str)
	if err == nil {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(0, pc)
	fn := runtime.FuncForPC(pc[2])
	file, line := fn.FileLine(pc[2])

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	t.Fatalf("\n*** Set(%q, %q), failed:\n    %q\n*** From before: %s:%d\n",
		key, val, err, file, line)
}

// Not used - think about removing
func ValidateGet(t *testing.T, cfg *Config, key string, exp interface{}) {
	str := fmt.Sprintf("%v", exp)
	val, err := cfg.Get(key)
	if err == nil && val == str {
		return
	}

	pc := make([]uintptr, 10)
	runtime.Callers(0, pc)
	fn := runtime.FuncForPC(pc[2])
	file, line := fn.FileLine(pc[2])

	if i := strings.LastIndex(file, "/"); i >= 0 {
		file = file[i+1:]
	}

	t.Fatalf("\n*** Get(%s) = %q, expected %q\n*** From before: %s:%d\n",
		key, val, exp, file, line)
}
