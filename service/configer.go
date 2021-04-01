package service

import (
	"dispatch-up/model"
	"github.com/go-yaml/yaml"
	"io/ioutil"
)

func InitConfigFromFile(path string) model.Config {
	var setting model.Config
	config, _ := ioutil.ReadFile(path)
	_ = yaml.Unmarshal(config, &setting)
	return setting
}
