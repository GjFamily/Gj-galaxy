package Gj_galaxy

import (
	"fmt"
	"reflect"

	"github.com/facebookgo/inject"
)

type ServerProvide interface {
	OnCreate() error // 初始化提供对象，在所有对象注册后执行
	OnStart() error  // 执行该服务对象
	OnStop() error   // 停止服务
}

type ServerProvideWrap struct {
	server ServerProvide
	name   string
	config interface{}
}

var (
	serverGraph inject.Graph
	servers     []ServerProvideWrap
	serverName  map[string]int
	inited      bool
)

func init() {
	serverGraph = inject.Graph{}
	servers = make([]ServerProvideWrap, 0)
	serverName = make(map[string]int)
	inited = false
}

func RegisterProvide(provide interface{}) error {
	return serverGraph.Provide(
		&inject.Object{Value: provide},
	)
}

func RegisterServer(server ServerProvide, name string) error {
	if inited {
		return fmt.Errorf("server is running, do not register server")
	}

	_, ok := serverName[name]
	if ok {
		return fmt.Errorf("server is exists, need checked again")
	}

	err := RegisterProvide(server)
	if err != nil {
		return err
	}
	wrap := ServerProvideWrap{server, name, nil}
	t := reflect.TypeOf(server).Elem()
	f, ok := t.FieldByName("Config")
	if ok {
		c := reflect.New(f.Type.Elem()).Interface()
		config.attachConfig(c, name)
		wrap.config = c
	}
	servers = append(servers, wrap)
	serverName[name] = len(servers)
	return nil
}

func RegisterConfig(config *Config) error {
	if inited {
		return fmt.Errorf("server is running, do not register config")
	}
	for index := 0; index < len(servers); index++ {
		wrap := servers[index]
		if wrap.config == nil {
			break
		}
		r := config.readConfig(wrap.name).(map[string]interface{})
		if r == nil {
			break
		}
		v := reflect.ValueOf(wrap.config).Elem()
		t := reflect.TypeOf(wrap.config).Elem()
		for k := 0; k < t.NumField(); k++ {
			key := t.Field(k).Tag.Get("json")
			logger.Debug(key)
			v.Field(k).Set(reflect.ValueOf(r[key]))
		}
		logger.Infof("[ Server ] config inject:%s -> %s", wrap.name, wrap.config)
		RegisterProvide(wrap.config)
	}
	return nil
}

func bindServer() error {
	err := serverGraph.Populate()
	if err != nil {
		return err
	}
	inited = true
	for index := 0; index < len(servers); index++ {
		wrap := servers[index]
		logger.Infof("[ Server ] bind %s -> onCreate", wrap.name)
		err = wrap.server.OnCreate()
		if err != nil {
			break
		}
	}
	return err
}

func StartServer() error {
	if !inited {
		return fmt.Errorf("server is not init, do not start")
	}
	l := len(servers) - 1
	for index := l; index >= 0; index-- {
		wrap := servers[index]
		logger.Infof("[ Server ] start %s -> onStart", wrap.name)
		err := wrap.server.OnStart()
		if err != nil {
			return err
		}
	}
	return nil
}

func StopServer() error {
	l := len(servers) - 1
	for index := l; index >= 0; index-- {
		wrap := servers[l]
		logger.Infof("[ Server ] stop %s -> onStop", wrap.name)
		err := wrap.server.OnStop()
		if err != nil {
			return err
		}
	}
	return nil
}
