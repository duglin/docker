package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/config"
	"github.com/docker/docker/pkg/homedir"
	flag "github.com/docker/docker/pkg/mflag"

	"github.com/docker/docker/utils"
)

type AuthConfig struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Auth          string `json:"auth"`
	Email         string `json:"email"`
	ServerAddress string `json:"serveraddress,omitempty"`
}

type ClientConfig struct {
	config.JsonConfig
	DockerHost  string            `json:",omitempty"`
	Registries  []AuthConfig      `json:",omitempty"`
	HttpHeaders map[string]string `json:",omitempty"`
}

const (
	DEFAULT_CONFIG_FILE = ".dockercli"
)

// 'docker config': get/set CLI config properties
func (cli *DockerCli) CmdConfig(args ...string) error {
	cmd := cli.Subcmd("config", "get|set|list|dump [[property] [value]]",
		"Manage the config file properties", true)

	cmd.Require(flag.Min, 1)

	utils.ParseFlags(cmd, args, false)

	return ProcessConfigCmdLine(cli, args)
}

func NewClientConfig(file string) (*ClientConfig, error) {
	if file == "" {
		file = filepath.Join(homedir.Get(), DEFAULT_CONFIG_FILE)
	}

	cc := &ClientConfig{}
	cc.MakeConfig(cc)
	cc.SetFile(file)

	if err := cc.Load(); err != nil {
		if !strings.Contains(err.Error(), "no such file") {
			return cc, err
		}
	}

	return cc, nil
}

// Should be temporary until we get rid of the old config file
func ConfigFileExists(file string) bool {
	if file == "" {
		file = filepath.Join(homedir.Get(), DEFAULT_CONFIG_FILE)
	}

	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

func ProcessConfigCmdLine(cli *DockerCli, args []string) error {
	switch strings.ToLower(args[0]) {
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("'list' doesn't have any parameters")
		}
		data, err := cli.config.List()
		if err != nil {
			return err
		}
		for k, v := range data {
			fmt.Fprintf(cli.out, "%s %s\n", k, v)
		}

	/* Saved in case we decide we want it later
	case "keys":
		for _, k := range cli.config.Keys() {
			fmt.Fprintf(cli.out, "%s\n", k)
		}
	*/

	case "get":
		if len(args) != 2 {
			return fmt.Errorf("'get' requires exactly one parameter")
		}
		val, err := cli.config.Get(args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(cli.out, "%s\n", val)

	case "set":
		if len(args) != 3 {
			return fmt.Errorf("'set' requires exactly two parameters")
		}
		err := cli.config.Set(args[1], args[2])
		if err != nil {
			return err
		}
		return cli.config.Save()

	case "dump":
		if len(args) != 1 {
			return fmt.Errorf("'list' doesn't have any parameters")
		}
		data, err := cli.config.Dump()
		if err != nil {
			return err
		}
		fmt.Fprintf(cli.out, "%s\n", data)

	default:
		return fmt.Errorf("Unknown config command: %s", args[0])
	}
	return nil
}
