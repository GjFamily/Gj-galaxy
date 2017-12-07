package socket

import (
	"fmt"
	"strings"
	"sync"
)

type Middleware func(socket Socket, next Next)

type EventHandler func(params ...interface{})

type Namespace interface {
	On(event string, handler EventHandler) error
	Use(middleware Middleware) Namespace
	Of(path string) (Namespace, error)
}

// 用于处理事件分发，逻辑业务管理
type namespace struct {
	e             *engine
	name          string
	fns           []Middleware
	sockets       map[string]Socket
	clientChannel chan *client
	events        map[string]*caller
	evMu          sync.Mutex
}

var (
	inlineEvent = []string{"connecting", "connect", "disconnecting", "disconnect"}
)

func newNamespace(e *engine, name string) *namespace {
	n := namespace{
		e,
		name,
		[]Middleware{},
		make(map[string]Socket),
		make(chan *client),
		make(map[string]*caller),
		sync.Mutex{},
	}
	n.listenChannel()
	return &n
}

func (n *namespace) Use(middleware Middleware) Namespace {
	n.fns = append(n.fns, middleware)
	return n
}

func (n *namespace) On(event string, handler EventHandler) error {
	c, err := newCaller(handler)
	if err != nil {
		return err
	}
	n.evMu.Lock()
	n.events[event] = c
	n.evMu.Unlock()
	return nil
}

func (n *namespace) Of(path string) (Namespace, error) {
	if path == "" || path == "/" {
		return n, nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = n.name + path
	if ns, ok := n.e.nss[path]; ok {
		return ns, fmt.Errorf("[ SOCKET ] namespace %s already used", path)
	}
	ns := newNamespace(n.e, path)
	n.e.nss[path] = ns
	return ns, nil
}

func (n *namespace) listenChannel() {
	go func() {
		for {
			select {
			case client := <-n.clientChannel:
				n.add(client)
			}
		}
	}()
}

func (n *namespace) run(socket Socket, c <-chan error, next Next) {
	l := len(n.fns)
	if l == 0 {
		next(nil)
		return
	}
	d := func(index int) {}
	var e error
	run := func(index int) {
		n.fns[index](socket, func(err error) AsyncResult {
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

func (n *namespace) add(client *client) <-chan *socketInline {
	socket := &socketInline{client: client}
	c := make(chan error)
	s := make(chan *socketInline)
	n.run(socket, c, func(err error) AsyncResult {
		if err != nil {
			c <- err
			return c
		}
		execute := func() {
			err = n.emit("connecting", socket, nil)
			if err != nil {
				c <- err
				return
			}
			n.sockets[socket.client.SID] = socket
			socket.onConnect()
			err = n.emit("connect", socket, nil)
			if err != nil {
				c <- err
				return
			}
			s <- socket
			c <- nil
		}
		go execute()
		return c
	})
	return s
}

func (n *namespace) remove(client *client, reason string) {
	socket, ok := n.sockets[client.SID]
	if !ok {
		logger.Debugf("[ SOCKET ] remove not in namespace:%s", n.name)
	}
	n.emit("disconnecting", socket, reason)
	delete(n.sockets, client.SID)
	n.emit("disconnect", socket, reason)
}

func (n *namespace) emit(event string, socket Socket, params ...interface{}) error {

	n.evMu.Lock()
	c, ok := n.events[event]
	n.evMu.Unlock()
	if !ok {
		return nil
	}
	args := c.GetArgs()

	retV := c.Call(socket, args)
	if len(retV) == 0 {
		return nil
	}
	return nil
}
