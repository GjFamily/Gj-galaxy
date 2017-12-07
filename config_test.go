package Gj_galaxy

import "testing"
import "strings"

type ConfigExample struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func TestConfig_LoadConfig(t *testing.T) {
	c := Config{}
	c.attachConfig(ConfigExample{}, "example")
	c.LoadConfig("config.json")
	sub := c.readConfig("example")
	if sub == nil {
		t.Fatalf("example config is error")
	}
}

func TestConfig_LoadFromReader(t *testing.T) {
	c := Config{}
	reader := strings.NewReader("{}")
	c.LoadFromReader(reader)
}
