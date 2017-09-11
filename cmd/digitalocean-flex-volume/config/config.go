package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
)

const (
	tokenFileEnv         = "DIGITALOCEAN_TOKEN_FILE_PATH"
	tokenEnv             = "DIGITALOCEAN_TOKEN"
	tokenDefaultLocation = "/etc/kubernetes/digitalocean.json"
)

// GetDigitalOceanToken uses environment variables to locate a Digital Ocean
// token. It will look at a file defined at en environment variable fisrt,
// then to an environment variable
func GetDigitalOceanToken() (string, error) {

	// try to load from file from env
	if f, ok := os.LookupEnv(tokenFileEnv); ok && f != "" {
		token, err := ReadTokenFromJSONFile(f)
		if err == nil && token != "" {
			return token, nil
		}
		glog.Infof("Could not find a valid configuration file at %s", f)
	}

	// try to load from environment
	if t, ok := os.LookupEnv(tokenEnv); ok {
		token := strings.TrimSpace(t)
		if token != "" {
			return token, nil
		}
		glog.Infof("Could not find a valid token at environment variable %s", tokenEnv)
	}

	//try the default location
	token, err := ReadTokenFromJSONFile(tokenDefaultLocation)
	if err == nil && token != "" {
		return token, nil
	}
	glog.Infof("Could not find a valid configuration file at %s", tokenDefaultLocation)

	return "", fmt.Errorf("No valid Digital Ocean tokens were found: %s", err)
}

// Config contains Digital Ocean configuration items
type Config struct {
	Token string `json:"token"`
}

// ReadTokenFromJSONFile reads the Digital Ocean token from a config file
func ReadTokenFromJSONFile(file string) (string, error) {
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	config := &Config{}
	err = json.Unmarshal(c, config)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(config.Token), nil

}
