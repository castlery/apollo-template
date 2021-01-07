package main

import (
	"errors"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	"log"
	"os"
)

// Config is used to configure Consul Template
type Config struct {
	apollo_address string
	// templateArr is the list of templates.
	templateArr []TemplateConfig `mapstructure:"templateArr"`
}

type TemplateConfig struct {
	Source      string
	Destination string
	MissKeyError bool
}

func ParseFile(path string) (*Config, error) {
	if path[0] != os.PathSeparator {
		// 相对路径
		cwd, _ := os.Getwd()
		path = cwd + string(os.PathSeparator) + path
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return Parse(string(bytes))
}

func Parse(s string) (*Config, error) {
	var shadow interface{}
	if err := hcl.Decode(&shadow, s); err != nil {
		return nil, err
	}

	// Convert to a map and flatten the keys we want to flatten
	parsed, ok := shadow.(map[string]interface{})
	if !ok {
		return nil, errors.New("error converting config")
	}

	var c Config
	// valid
	if _, ok := parsed["template"].([]map[string]interface{}); !ok {
		log.Fatal("must set templateArr config")
	}
	// convert to config
	var arr []TemplateConfig
	for _, t := range parsed["template"].([]map[string]interface{}) {
		config := TemplateConfig{
			Source:      t["source"].(string),
			Destination: t["destination"].(string),
			MissKeyError: t["error_on_missing_key"].(bool),
		}
		arr = append(arr, config)
	}
	c.templateArr = arr
	return &c, nil
}
