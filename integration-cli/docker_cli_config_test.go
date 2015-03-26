package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/docker/docker/pkg/homedir"
)

func TestConfigHttpHeader(t *testing.T) {
	var headers map[string][]string

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			headers = r.Header
		}))
	defer server.Close()

	homeKey := homedir.Key()
	homeVal := homedir.Get()
	tmpDir, _ := ioutil.TempDir("", "fake-home")
	tmpCfg := filepath.Join(tmpDir, ".dockercfg")

	defer func() { os.Setenv(homeKey, homeVal) }()
	os.Setenv(homeKey, tmpDir)

	data := `{
	    "auths": {},
		"httpHeaders": {
			"MyHeader": "MyValue"
		}
	}`

	ioutil.WriteFile(tmpCfg, []byte(data), 0600)

	cmd := exec.Command(dockerBinary, "-H="+server.URL[7:], "ps")
	out, _, _ := runCommandWithOutput(cmd)

	if headers["Myheader"] == nil || headers["Myheader"][0] != "MyValue" {
		t.Fatal("Missing/bad header: %q\nout:%v", headers, out)
	}

	logDone("config - add new http headers")
}
