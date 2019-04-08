package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"text/template"

	"github.com/bhmj/jsonslice"
	"golang.org/x/crypto/bcrypt"
)

type jsonmap = map[string]interface{}

const logFlags = log.LstdFlags | log.Lshortfile | log.LUTC

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(logFlags)

	configTemplate, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("unable to read template: %v", err)
	}

	funcMap := template.FuncMap{
		"env":      os.Getenv,
		"hostname": os.Hostname,
		"bcrypt":   hashBcrypt,
		"app":      getJSONSlice("PLATFORM_APPLICATION"),
		"var":      getJSONSlice("PLATFORM_VARIABLES"),
		"route":    getRoutes(),
		"rel":      getJSONSlice("PLATFORM_RELATIONSHIPS"),
		"slice": func(data, path string) (string, error) {
			sub := returnSlice([]byte(data))
			return sub(path)
		},
		"json": func(data string) (interface{}, error) {
			var v interface{}
			err = json.Unmarshal([]byte(data), &v)
			return v, err
		},
	}

	tpl, err := template.New("config.yml").Funcs(funcMap).Parse(string(configTemplate))
	if err != nil {
		log.Fatalf("unable to compile template: %v", err)
	}

	buffer := &bytes.Buffer{}
	if err := tpl.Execute(buffer, nil); err == nil {
		io.Copy(os.Stdout, buffer)
	} else {
		log.Fatalf("unable to render template: %v", err)
	}
}

func hashBcrypt(password string) (string, error) {
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(passwordBytes), nil
}

func getJSONSlice(key string) func(string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return returnError(fmt.Errorf("environment variable %s not found", key))
	}

	jsonValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return returnError(err)
	}

	return returnSlice(jsonValue)
}

func getRoutes() func(string) (string, error) {
	value, ok := os.LookupEnv("PLATFORM_ROUTES")
	if !ok {
		return returnError(errors.New("environment variable PLATFORM_ROUTES not found"))
	}

	jsonValue, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return returnError(err)
	}

	var routes jsonmap
	err = json.Unmarshal(jsonValue, &routes)
	if err != nil {
		return returnError(err)
	}

	var result = make([]jsonmap, 0, len(routes))
	for k, v := range routes {
		m, ok := v.(jsonmap)
		if !ok {
			return returnError(fmt.Errorf("invalid value for %s: %#v", k, v))
		}

		m["url"], err = url.Parse(k)
		if err != nil {
			return returnError(err)
		}

		result = append(result, m)
	}

	jsonValue, err = json.Marshal(result)
	if err != nil {
		return returnError(err)
	}

	return returnSlice(jsonValue)
}

func returnError(err error) func(string) (string, error) {
	return func(string) (string, error) {
		return "", err
	}
}

func returnSlice(data []byte) func(string) (string, error) {
	return func(path string) (string, error) {
		v, err := jsonslice.Get(data, path)
		return string(v), err
	}
}
