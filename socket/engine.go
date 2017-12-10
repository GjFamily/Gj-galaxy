package socket

import (
	"net"
	"net/http"
)

type AsyncResult <-chan error
type Next func(err error) AsyncResult

type HttpMux interface {
	Handle(pattern string, handler http.Handler)
}

type Logger interface {
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
}

type Engine interface {
	Listen() error
	Attach(http HttpMux) error
	BindTCP(tcp string) error
	BindUDP(udp string) error
	Close() error
	Of(path string) (Namespace, error)
}

type engine struct {
	Logger Logger

	core Core

	TCPAddr *net.TCPAddr
	UDPAddr *net.UDPAddr
	WebPath string

	certFile string
	keyFile  string

	nss map[string]*namespace

	stop      chan bool
	listening bool
}

func NewEngine(options map[string]interface{}) (Engine, error) {
	var e = &engine{}
	webPath, ok := options["webPath"]
	if ok {
		e.WebPath = webPath.(string)
	} else {
		e.WebPath = "/galaxy.socket"
	}
	e.nss["/"] = newNamespace(e, "/")
	e.core = newCore(e)
	e.stop = make(chan bool)
	e.listening = false
	return e, nil
}

func (e *engine) SetLogger(logger Logger) {
	e.Logger = logger
}

func (e *engine) Listen() error {
	err := e.core.Listen(e.TCPAddr, e.UDPAddr)
	if err != nil {
		return err
	}
	go func() {
		e.listening = true
		for {
			select {
			case <-e.stop:
				break
			}
		}
		e.listening = false
	}()
	return nil
}

func (e *engine) Attach(http HttpMux) error {
	return e.core.Attach(e.WebPath, http)
}

func (e *engine) BindTCP(tcp string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcp)
	if err != nil {
		return err
	}
	e.TCPAddr = tcpAddr
	return nil
}

func (e *engine) BindUDP(udp string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", udp)
	if err != nil {
		return err
	}
	e.UDPAddr = udpAddr
	return nil
}

func (e *engine) Close() error {
	err := e.core.Close()
	if err != nil {
		return err
	}
	if e.listening {
		e.stop <- true
	}
	return nil
}

func (e *engine) Of(path string) (Namespace, error) {
	return e.nss["/"].Of(path)
}
