package room

import (
	"Gj-galaxy/socket"

	"github.com/Sirupsen/logrus"
)

type Server struct {
	Socket *socket.Server `inject:""`
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
