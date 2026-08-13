package sbcli

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Credentials struct {
	Host              string `json:"host"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	SkipSslValidation bool   `json:"skip_ssl_validation"`
}

type Config struct {
	Credentials
	OrganizationGUID string `json:"organization_guid,omitempty"`
	SpaceGUID        string `json:"space_guid,omitempty"`
}

const (
	ConfigFile = ".sb"
)

// configPath returns the absolute path to the .sb config file in the
// user's home directory. The config lives in $HOME so it follows the
// user across shells and working directories, matching how `cf` and
// similar CLIs persist their session state.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigFile), nil
}

func (c *Config) load() error {
	if os.Getenv("SB_HOST") != "" &&
		os.Getenv("SB_USERNAME") != "" &&
		os.Getenv("SB_PASSWORD") != "" {
		c.Credentials.Host = CleanTargetURI(os.Getenv("SB_HOST"))
		c.Credentials.Username = os.Getenv("SB_USERNAME")
		c.Credentials.Password = os.Getenv("SB_PASSWORD")
		if skipVerify, ok := os.LookupEnv("SB_SKIP_SSL_VERIFY"); ok {
			value, err := strconv.ParseBool(skipVerify)
			if err != nil {
				log.Fatal(err)
			}
			c.SkipSslValidation = value
		}
		return nil
	}

	file, err := configPath()
	if err != nil {
		return err
	}

	jsonFile, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("config: config not found")
		}
		return err
	}

	return json.Unmarshal(jsonFile, c)
}

func (c *Config) save() error {
	configJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}
	file, err := configPath()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, configJSON, 0600)
}

func LoadConfig() *Config {
	c := Config{}
	c.load()
	return &c
}
