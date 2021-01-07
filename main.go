package main

import (
	"flag"
	"github.com/philchia/agollo/v4"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
)

var apolloAddress string
var once sync.Once

func main() {
	// command line flag
	var configPath string
	var apollo string
	flag.StringVar(&configPath, "config", "config.hcl", "配置文件地址，默认在当前目录下找config.hcl")
	if flag.Lookup(apollo)!=nil {
		flag.StringVar(&apollo, "apollo", "", "apollo config service 的地址")
	}
	flag.Parse()
	if strings.EqualFold("", apollo) {
		apolloAddress = os.Getenv("APOLLO_CONFIG_SERVICE_ADDRESS")
		if strings.EqualFold("", apolloAddress) {
			log.Fatal("apollo address must be set")
		}
	} else {
		apolloAddress = apollo
	}

	// load templateArr config
	config := loadTemplateConfig(configPath)

	for _, templateConfig := range config.templateArr {
		// templateArr render
		tplStr := loadTemplate(templateConfig.Source)
		configMap := map[string]interface{}{}
		t, err := template.New(templateConfig.Source).Funcs(*enhanceTemplate(configMap)).Parse(tplStr)
		if err != nil {
			log.Fatal(err)
		}
		if templateConfig.MissKeyError{
			t.Option("missingkey=error")
		}
		f, err := os.Create(templateConfig.Destination)
		err = t.Execute(f, nil)
		if err != nil {
			log.Fatalf("execution failed: %s", err)
		}
		f.Close()
	}
}

func loadConfigFromApollo(app string,cluster string, namespaces string) map[string]interface{} {
	namespaceArr := strings.Split(namespaces, ",")
	log.Println("connection to apollo: ", apolloAddress)
	curPath, _ := os.Getwd()
	cachePath := curPath + string(os.PathSeparator) + "cache"
	once.Do(func() {
		err := agollo.Start(&agollo.Conf{
			AppID:          app,
			Cluster:        cluster,
			NameSpaceNames: namespaceArr,
			MetaAddr:       apolloAddress,
			CacheDir:       cachePath,
		})
		if err != nil {
			log.Fatal(err)
		}
	})
	configMap := map[string]interface{}{}
	for _, ns := range namespaceArr {
		content := agollo.GetPropertiesContent(agollo.WithNamespace(ns))
		items := strings.Split(content, "\n")
		for _, item := range items {
			if "" != item {
				arr := strings.Split(item, "=")
				configMap[arr[0]] = arr[1]
			}
		}
	}
	return configMap
}

func loadTemplate(path string) (content string) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("execution failed: %s", err)
	}
	content = string(bytes)
	return content
}

func loadTemplateConfig(path string) *Config {
	configPtr, err := ParseFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return configPtr
}
