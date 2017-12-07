package socket

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type ExitStatus int

const (
	_ ExitStatus = iota
	CLIENT
	SERVER
)

type Component interface {
	ID() string
	Filter(message string) bool
	Remove(socket Socket)
}

// 用于处理请求信息，面向交互
type Socket interface {
	Close()                                         // 关闭链接
	Disconnect()                                    // 断开namespace
	Error(err error)                                // 发送错误包
	Emit(event string, params ...interface{}) error // 触发事件
	//On(event string, handler EventHandler) error    // 监听事件

	Add(component Component) error    // 添加组件
	Delete(component Component) error // 移除组件
	Group(message string) []Component // 返回分组组件
}

type socketInline struct {
	client     *client
	namespace  *namespace
	ExitStatus ExitStatus
	fns        []Middleware
	components map[string]Component
	connected  bool

	acksmu sync.Mutex
	acks   map[int]*caller
	ids    int

	events map[string]*caller
	evMu   sync.Mutex
}

func newSocket(namespace *namespace, client *client) Socket {
	s := socketInline{
		client,
		namespace,
		nil,
		[]Middleware{},
		make(map[string]Component),
		true,
		sync.Mutex{},
		make(map[int]*caller),
		0,
		make(map[string]*caller),
		sync.Mutex{},
	}
	return &s
}

func (s *socketInline) Close() {
	s.ExitStatus = SERVER
	s.client.Close()
}

func (s *socketInline) Disconnect() {
	logger.Debugf("")
	s.ExitStatus = SERVER

	s.packet(&packet{Type: P_DISCONNECT}, nil)
	s.onClose("server namespace disconnect")
}

func (s *socketInline) Error(err error) {
	s.packet(&packet{Data: err}, nil)
}

func (s *socketInline) Emit(event string, args ...interface{}) error {
	packet := &packet{
		Type: P_EVENT,
	}
	var c *caller
	if l := len(args); l > 0 {
		fv := reflect.ValueOf(args[l-1])
		if fv.Kind() == reflect.Func {
			var err error
			c, err = newCaller(args[l-1])
			if err != nil {
				return err
			}
			args = args[:l-1]
			s.acksmu.Lock()
			s.acks[s.ids] = c
			packet.Id = s.ids + 1
			s.acksmu.Unlock()
		}
	}
	packet.Data = args
	s.packet(packet, nil)
	return nil
}

func (s *socketInline) Add(component Component) error {
	id := component.ID()
	if _, ok := s.components[id]; ok {
		return fmt.Errorf("Component %s is exits", id)
	}
	s.components[id] = component
	return nil
}

func (s *socketInline) Delete(component Component) error {
	id := component.ID()
	if _, ok := s.components[id]; !ok {
		return fmt.Errorf("Component %s is not exits", id)
	}
	delete(s.components, id)
	return nil
}

func (s *socketInline) Group(message string) []Component {
	c := make([]Component, 0)
	for _, component := range s.components {
		if component.Filter(message) {
			c = append(c, component)
		}
	}
	return c
}

func (s *socketInline) onPacket(packet *packet) {
	switch packet.Type {
	case P_EVENT:
		s.onEvent(packet)
	case P_ACK:
		s.onAck(packet)
	case P_DISCONNECT:
		s.onDisconnect(packet)
	case P_ERROR:
		s.onError(errors.New(packet.Data.(string)))
	}
}

func (s *socketInline) Use(middleware Middleware) Socket {
	s.fns = append(s.fns, middleware)
	return s
}

func (s *socketInline) run(c <-chan error, next Next) {
	l := len(s.fns)
	if l == 0 {
		next(nil)
		return
	}
	d := func(index int) {}
	var e error
	run := func(index int) {
		s.fns[index](s, func(err error) AsyncResult {
			e = err
			if err != nil {
				return next(err)
			}
			if index+1 >= l {
				return next(nil)
			} else {
				d(index + 1)
			}
			c <- e
			return c
		})
	}
	d = run
	go run(0)
}

func (s *socketInline) dispatch(event string, args []interface{}) {
	c := make(chan error)
	s.run(c, func(err error) AsyncResult {
		if err != nil {
			c <- err
			s.Error(err)
			return c
		}
		execute := func() {
			s.namespace.emit(event, s, args...)
			c <- nil
		}
		go execute()
		return c
	})
}

func (s *socketInline) onEvent(packet *packet) {
	var args = packet.Data.([]interface{})
	if packet.Id != -1 {
		args = append(args, s.ack(packet.Id))
	}
	event := args[0].(string)
	args = args[1:]

	s.dispatch(event, args)
}

func (s *socketInline) onDisconnect(packet *packet) {
	s.ExitStatus = CLIENT
	s.onClose("client namespace disconnect")
}

func (s *socketInline) onClose(reason string) {
	if !s.connected {
		return
	}
	s.connected = false
	s.namespace.remove(s.client, reason)
	for _, component := range s.components {
		component.Remove(s)
	}
	s.client.Remove(s)
}

func (s *socketInline) onError(err error) {
	s.namespace.emit("error", s, err)
}

func (s *socketInline) onAck(packet *packet) {
	var ack = s.acks[packet.Id]
	fv := reflect.ValueOf(ack)
	if fv.Kind() == reflect.Func {
		ack.Apply(s, packet.Data)
		delete(s.acks, packet.Id)
	} else {
		logger.Debugf("bad ack %s", packet.Id)
	}
}

func (s *socketInline) ack(id int) func(args ...interface{}) {
	var sent = false
	return func(args ...interface{}) {
		// prevent double callbacks
		if sent {
			return
		}
		s.packet(&packet{
			Id:   id,
			Type: P_ACK,
			Data: args,
		}, nil)

		sent = true
	}
}

func (s *socketInline) onConnect() {
	s.packet(&packet{Type: P_CONNECT}, nil)
}

func (s *socketInline) packet(packet *packet, opts *packetOpt) {
	packet.NSP = s.namespace.name
	if opts == nil {
		opts = &packetOpt{}
	}
	opts.compress = false != opts.compress
	s.client.Packet(packet, opts)
}
