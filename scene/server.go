package scene

import (
	"Gj-galaxy/platform"
	"Gj-galaxy/room"

	"github.com/Sirupsen/logrus"
)

type Server struct {
	Room     *room.Server     `inject:""`
	Platform *platform.Server `inject:""`
}

var (
	logger *logrus.Logger
)

func init() {
	logger = logrus.StandardLogger()
}
func (server *Server) OnCreate() error {
	return nil
}

func (server *Server) OnStart() error {
	return nil
}

func (server *Server) OnStop() error {
	return nil
}

func (server *Server) GetConfig() interface{} {
	return nil
}
