package main

import (
	"encoding/base64"
	"os"
	"text/template"

	"github.com/pkg/errors"
)

func enhanceTemplate(configMap map[string]interface{}) *template.FuncMap {
	return &template.FuncMap{
		"apollo": apolloFunc(),
		// Helper functions
		"base64Decode": base64Decode,
		"base64Encode": base64Encode,
		"env":          envFunc,
	}
}

func envFunc(env string) string {
	return os.Getenv(env)
}

func base64Decode(s string) (string, error) {
	v, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.Wrap(err, "base64Decode")
	}
	return string(v), nil
}

// base64Encode encodes the given value into a string represented as base64.
func base64Encode(s string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func apolloFunc() func(string, string) interface{} {
	return func(app string, ns string) interface{} {
		cluster := "default"
		configMap := loadConfigFromApollo(app, cluster, ns)
		data := struct {
			Data map[string]interface{}
		}{Data: configMap}
		return data
	}
}
