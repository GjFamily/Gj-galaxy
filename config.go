package Gj_galaxy

import (
	"encoding/json"
	"io"
	"os"
)

// ConfigFile config.json file info
type Config struct {
	Db        DbConfig               `json:"db"`
	Redis     RedisConfig            `json:"redis"`
	Advertise string                 `json:"advertise"`
	Server    map[string]interface{} `json:"server"`
}

type DbConfig struct {
	Dsn string `json:"url"`
}

type RedisConfig struct {
	Dsn string `json:"url"`
}

func (config *Config) attachConfig(subConfig interface{}, name string) {
	if config.Server == nil {
		config.Server = make(map[string]interface{})
	}
	config.Server[name] = subConfig
}

func (config *Config) readConfig(name string) interface{} {
	if config.Server == nil {
		return nil
	}
	v, ok := config.Server[name]
	if !ok {
		return nil
	}
	return v
}

func (config *Config) LoadFromReader(configData io.Reader) error {
	if err := json.NewDecoder(configData).Decode(&config); err != nil {
		return err
	}
	return nil
}

func (config *Config) LoadConfig(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		err = config.LoadFromReader(file)
		return err
	} else {
		return err
	}
	return nil
}
