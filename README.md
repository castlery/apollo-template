## install
1. 根据系统下载release 包
2. 解析后即可使用

## usage

1. 使用命令行指定 apollo-config
    > apollo-template -config=/path/to/config.hcl -apollo=apollo-config.address

2. 从环境变量中获取 apollo-config

    > export APOLLO_CONFIG_SERVICE_ADDRESS=http://apollo-config:8080 
                              
    > apollo-template -config=/path/to/config.hcl

### arguments

> -config: 配置文件地址，必须指定。格式完全兼容consul-template，

> -apollo: apollo-config-service 地址。推荐设置环境 APOLLO_CONFIG_SERVICE_ADDRESS 来配置 

## template render

确定一个配置的获取地址需要 appId(project name)，namespace,cluster 三个参数，其中cluster 我们目前不用，
所以使用apollo的默认值"default"，appId 和 namespace 可以在模板文件中进行配置.我们目前的用法下：
appId 等价于 project name，namespace 等价于 部署环境


```yaml
{{ $namespace := printf "%s" (env "APP_ENV") }}
{{ with apollo "appId" $namespace }}

spring.datasource:
  url: {{ .Data.horus_db_url }}
  username: {{ .Data.horus_db_username }}
  password: {{ .Data.horus_db_password }}
eia:
  url: {{ .Data.eia_url }}

{{ end }}
```

## consul-template 兼容性

目前会解析的node如下所示：

```yaml
template {
  source      = "/Users/caipeijun/go/src/apollo-go/tpl/application-default.yml.tmpl"
  destination = "/Users/caipeijun/go/src/apollo-go/tpl/application-default.yml"
  error_on_missing_key = true
}

template {
  source      = "/Users/caipeijun/go/src/apollo-go/tpl/sentry.properties.tmpl"
  destination = "/Users/caipeijun/go/src/apollo-go/tpl/sentry.properties"
  error_on_missing_key = true
}

其他的node仍然可以存在只是不会起作用了。

```

## 迁移指南

1. 配置从vault 迁移到 apollo

    todo confluence 文档

2. 修改项目打包文件

    以java为例，项目的打包配置目录如下
```
├── etc             
│   ├── consul-template.hcl    // 模板渲染的配置文件 
│   ├── docker-build-assets.sh // 在docker 构建中会在镜像中执行的脚本文件，用来执行模板渲染操作
│   └── templates
│       ├── application-default.yml.tmpl // 配置文件模板
│       └── sentry.properties.tmpl
├── Dockerfile      // docker 打包文件 
```    

a. 修改Dockerfile，替换vault相关的工具，使用apollo-template

将
```shell script
RUN curl -o /run/vault.zip -fSL "https://releases.hashicorp.com/vault/0.7.3/vault_0.7.3_linux_amd64.zip" \
    && unzip -d /usr/local/bin/ /run/vault.zip \
    && rm /run/vault.zip

RUN curl -o /run/consul-template.zip -fSL "https://releases.hashicorp.com/consul-template/0.19.5/consul-template_0.19.5_linux_amd64.zip" \
    && unzip -d /usr/local/bin/ /run/consul-template.zip \
    && rm /run/consul-template.zip
```
替换成
```shell script
RUN curl -o /run/apollo-template.tar.gz -fSL "https://github.com/castlery/apollo-template/releases/download/0.0.2/apollo-template_0.0.3_linux_amd64.tar.gz" \
    && tar -zxvf /run/apollo-template.tar.gz --directory /usr/local/bin/ \
    && rm /run/apollo-template.tar.gz
```

b. 修改 docker-build-assets.sh，替换vault,consul-template，使用apollo-template 渲染模版，替换之后的文件如下：

```shell script
#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

export APOLLO_CONFIG_SERVICE_ADDRESS=http://apollo-config-service-prod.castlery.internal

echo "Waiting for apollo-template to refresh config files"
apollo-template -config "/app/etc/apollo-template.hcl"

```

c. 修改模板

将
```golang
{{ $secret_path := printf "secret/appId/%s" (env "APP_ENV") }}
{{ with secret $secret_path }}
```
替换成
```golang
{{ with apollo "appId" (env "APP_ENV") }}
```
