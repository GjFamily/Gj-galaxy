package Gj_galaxy

import (
	"testing"
)

type ServerProvideExample struct {
	Config *ServerProvideExampleConfig `inject:""`
}

type ServerProvideExampleConfig struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func (server *ServerProvideExample) OnCreate() error {
	logger.Debug("%s", server.Config)
	logger.Debug("create")
	return nil
}

func (server *ServerProvideExample) OnStart() error {

	logger.Debug("start")
	return nil
}

func (server *ServerProvideExample) OnStop() error {
	logger.Debug("stop")
	return nil
}

func TestContainer(t *testing.T) {
	var server ServerProvideExample
	errRegister := RegisterServer(&server, "example")
	if errRegister != nil {
		t.Fatal(errRegister)
	}

	config.LoadConfig("config.json")
	RegisterConfig(config)

	errBind := bindServer()
	if errBind != nil {
		t.Fatal(errBind)
	}
	errStart := StartServer()
	if errStart != nil {
		t.Error(errStart)
	}
	errStop := StopServer()
	if errStop != nil {
		t.Error(errStop)
	}

}
