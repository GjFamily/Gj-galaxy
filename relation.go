package Gj_galaxy

import (
	"Gj-galaxy/platform"
	"Gj-galaxy/restful"
	"Gj-galaxy/web"
	"database/sql"

	"Gj-galaxy/socket"
)

type PlatformServer struct {
	Platform platform.Platform
	DB       *sql.DB `inject:""`
}

func (server *PlatformServer) OnCreate() error {
	return nil
}

func (server *PlatformServer) OnStart() error {
	return nil
}

func (server *PlatformServer) OnStop() error {
	return nil
}

func (server *PlatformServer) GetConfig() interface{} {
	return nil
}

type WebServer struct {
	Root web.Root
	Web  web.Web
	//Redis     *redis.Pool   `inject:""`
	Config *WebConfig `inject:""`
}

type WebConfig struct {
	Dns string `json:"url"`
}

func (server *WebServer) OnCreate() error {
	root := web.NewRoot()
	webServer, err := web.NewWeb(server.Config.Dns, root)
	if err != nil {
		return err
	}
	server.Web = webServer
	server.Root = root
	return nil
}

func (server *WebServer) OnStart() error {
	return server.Web.Listen()
}

func (server *WebServer) OnStop() error {
	return nil
}

type SocketServer struct {
	Engine socket.Engine
	Web    web.Web       `inject:""`
	Config *SocketConfig `inject:""`
}

type SocketConfig struct {
	Tcp     string                 `json:"tcp"`
	Udp     string                 `json:"udp"`
	Options map[string]interface{} `json:"options"`
}

func (server *SocketServer) OnCreate() error {
	engine, err := socket.NewEngine(server.Config.Options)
	if err != nil {
		return err
	}
	err = engine.Attach(server.Web.GetRouter())
	if err != nil {
		return err
	}
	if server.Config.Tcp != "" {
		err = engine.BindTCP(server.Config.Tcp)
		if err != nil {
			return err
		}
	}
	if server.Config.Udp != "" {
		err = engine.BindUDP(server.Config.Udp)
		if err != nil {
			return err
		}
	}
	server.Engine = engine

	return nil
}

func (server *SocketServer) OnStart() error {
	return server.Engine.Listen()
}

func (server *SocketServer) OnStop() error {
	return server.Engine.Close()
}

type RestfulServer struct {
	Router   web.Router        `inject:""`
	Platform platform.Platform `inject:""`
}

func (server *RestfulServer) OnCreate() error {
	restful.NewRestful(server.Router)

	return nil
}

func (server *RestfulServer) OnStart() error {
	return nil
}

func (server *RestfulServer) OnStop() error {
	return nil
}
