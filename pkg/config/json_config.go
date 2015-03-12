package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type JsonConfig struct {
	Config
	file string
}

func (jcfg *JsonConfig) File() string {
	return jcfg.file
}

func (jcfg *JsonConfig) SetFile(fn string) error {
	jcfg.file = fn
	if fn != "" {
		if _, err := os.Stat(fn); err == nil {
			// If file doesn't exist then just skip trying to load it
			if err = jcfg.Load(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (jcfg *JsonConfig) Save() error {
	if jcfg.file == "" {
		return fmt.Errorf("Missing file")
	}
	data, err := jcfg.Dump()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(jcfg.file, []byte(data), 0600)
	if err != nil {
		return fmt.Errorf("Error writing file(%s): %s", jcfg.file, err.Error())
	}
	return nil
}

func (jcfg *JsonConfig) Load() error {
	if jcfg.file == "" {
		return fmt.Errorf("No file defined")
	}
	data, err := ioutil.ReadFile(jcfg.file)
	if err != nil {
		return fmt.Errorf("Error reading file(%s): %s", jcfg.file, err.Error())
	}
	err = json.Unmarshal(data, jcfg.data)
	return err
}
