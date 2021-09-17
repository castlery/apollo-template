package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/philchia/agollo/v4"
	"gopkg.in/yaml.v2"
)

var apolloAddress string
var once sync.Once

func main() {
	// command line flag
	var configPath string
	var apollo string
	flag.StringVar(&configPath, "config", "config.hcl", "配置文件地址，默认在当前目录下找config.hcl")
	flag.StringVar(&apollo, "apollo", "", "apollo config service 的地址")
	flag.Parse()
	if strings.EqualFold("", apollo) {
		apolloAddressFromEnv := os.Getenv("APOLLO_CONFIG_SERVICE_ADDRESS")
		if strings.EqualFold("", apolloAddressFromEnv) {
			log.Fatal("apollo address must be set")
		} else {
			apolloAddress = apolloAddressFromEnv
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
		if templateConfig.MissKeyError {
			t.Option("missingkey=error")
		}
		fileName := templateConfig.Destination
		dirName := filepath.Dir(fileName)
		err = os.MkdirAll(dirName, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		err = t.Execute(f, nil)
		if err != nil {
			log.Fatalf("execution failed: %s", err)
		}
		f.Close()
	}
}

func loadConfigFromApollo(app string, cluster string, namespaces string) map[string]interface{} {
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
		if strings.HasSuffix(ns, ".yaml") || strings.HasSuffix(ns, ".yml") {
			// yaml parse
			content := agollo.GetString("content", agollo.WithNamespace(ns))
			yaml.Unmarshal([]byte(content), configMap)
		} else {
			// properties parse
			content := agollo.GetPropertiesContent(agollo.WithNamespace(ns))
			items := strings.Split(content, "\n")
			for _, item := range items {
				if "" != item {
					arr := strings.Split(item, "=")
					configMap[arr[0]] = arr[1]
				}
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
