# Chrome Web Store Upload & Publish Tools
[![CircleCI](https://circleci.com/gh/anhnguyentb/cws-publish/tree/master.svg?style=svg)](https://circleci.com/gh/anhnguyentb/cws-publish/tree/master)

A CLI program to upload/publish extensions to the [Chrome Web Store](https://chrome.google.com/webstore/category/extensions).

## Installation
1. Installing with Docker
``docker pull anhnguyentb90/cws-publish``
2. Installing with development environment. Please make sure Go already installed
```
dep ensure -v
go run main.go
```

## Usage

Overview all of commands with CLI
```bash
Includes tools to build & publish Chrome Web Store

Usage:
  cws-publish [command]

Available Commands:
  build-store-configs Lookup store configs then copy into CWS provider folder
  help                Help about any command
  upload              Upload a zip file into CWS

Flags:
      --config string   config file (default is $HOME/.cws-publish.yaml)
  -h, --help            help for cws-publish
  -t, --toggle          Help message for toggle

Use "cws-publish [command] --help" for more information about a command.
```

### 1. Lookup store configs then copy into providers folder
```bash
Usage:
  cws-publish build-store-configs [flags]

Flags:
  -d, --dest string   Destination directory which be stored store provider
  -h, --help          help for build-store-configs
  -s, --src string    Source directory which be located store configs (YAML & Provider)

Global Flags:
      --config string   config file (default is $HOME/.cws-publish.yaml)
```

### 2. Upload & publish a zip file into CWS
```bash
Usage:
  cws-publish upload [flags]

Flags:
  -h, --help             help for upload
  -p, --publish          Publish CWS item immediately after zip file uploaded
  -t, --target string    Publish target (trustedTesters/default) (default "default")
  -z, --zipPath string   CWS zip file path

Global Flags:
      --config string   config file (default is $HOME/.cws-publish.yaml)
```

There are having 4 of environment variable should be defined (they are the credentials so should be defined as environment variable)

`EXTENSION_ID` - Your extension ID, get from CWS

`GOOGLE_CLIENT_ID` `GOOGLE_CLIENT_SECRET` `GOOGLE_REFRESH_TOKEN` please follow [at here](https://github.com/anhnguyentb/cws-publish/blob/master/How%20to%20generate%20Google%20API%20keys.md)