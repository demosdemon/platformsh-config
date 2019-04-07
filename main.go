package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/elgs/gojq"
	"golang.org/x/crypto/bcrypt"
)

type jsonMap = map[string]interface{}

type config struct {
	TokenURL    string
	RegistryURL string
}

func main() {
	configTemplate, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "E unable to read template: %v\n", err)
		os.Exit(4)
	}

	application, _ := getJSONVariable("PLATFORM_APPLICATION")
	variables, _ := getJSONVariable("PLATFORM_VARIABLES")
	routes, _ := getJSONVariable("PLATFORM_ROUTES")
	rels, _ := getJSONVariable("PLATFORM_RELATIONSHIPS")

	funcMap := template.FuncMap{
		"env":      os.Getenv,
		"hostname": os.Hostname,
		"app":      application.Query,
		"var":      variables.Query,
		"route":    routes.Query,
		"rel":      rels.Query,
		"bcrypt":   hashBcrypt,
	}

	tpl := template.New("config.yml")
	tpl = tpl.Funcs(funcMap)
	tpl, err = tpl.Parse(string(configTemplate))
	if err != nil {
		fmt.Fprintf(os.Stderr, "E unable to compile template: %v\n", err)
		os.Exit(1)
	}

	cfg := new(config)
	if routeMap, ok := routes.Data.(jsonMap); ok {
		for url, v := range routeMap {
			if vMap, ok := v.(jsonMap); ok {
				if upstream, ok := vMap["upstream"]; ok {
					if s, ok := upstream.(string); ok {
						switch s {
						case "auth":
							cfg.TokenURL = url
						case "registry":
							cfg.RegistryURL = url
						}
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "failed to convert vMap %s %#v\n", url, v)
			}
		}
	} else {
		fmt.Fprintf(os.Stderr, "failed to convert routeMap %#v\n", routes.Data)
	}

	err = tpl.Execute(os.Stdout, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "E unable to render template: %v\n", err)
		os.Exit(2)
	}
}

func hashBcrypt(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordBytes), nil
}

func getJSONVariable(key string) (*gojq.JQ, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return nil, fmt.Errorf("key %q not found", key)
	}

	jsonValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	return gojq.NewStringQuery(string(jsonValue))
}
