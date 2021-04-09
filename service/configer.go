package service

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

func Config(settings map[string]interface{}) {
	fmt.Println("config path: " + "~/.dispatch/dispatch.yaml")
	fmt.Println()
	b, err := yaml.Marshal(settings)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
}
