package socket

import (
	"fmt"
	"net"
	"sync"

	"Gj-galaxy/socket/protocol"
)

type Conn interface {
	Write([]byte)
	Accept() (int, <-chan []byte, <-chan string, <-chan error)
	UnAccept(id int)
	Disconnect()
	Session() string
}

type Core interface {
	Listen(tcpAddr *net.TCPAddr, udpAddr *net.UDPAddr) error
	Attach(path string, http HttpMux) error
	Accept() <-chan Conn
	Close() error
}

type core struct {
	WebSocket protocol.Protocol
	Tcp       protocol.Protocol
	Kcp       protocol.Protocol
	stop      chan bool
	conn      chan Conn
}

func newCore() Core {
	co := core{}

	return &co
}

func (co *core) Attach(path string, http HttpMux) error {
	s, err := protocol.WebSocketAttach(path, http)
	if err != nil {
		return err
	}
	co.WebSocket = s
	return nil
}

func (co *core) Listen(tcpAddr *net.TCPAddr, udpAddr *net.UDPAddr) error {
	if tcpAddr != nil {
		tcp, err := protocol.TcpListen(tcpAddr)
		if err != nil {
			return err
		}
		logger.Debugf("[ Socket ] listen Tcp :%s", tcpAddr)
		co.Tcp = tcp
	}
	if udpAddr != nil {
		kcp, err := protocol.KcpListen(udpAddr)
		if err != nil {
			return err
		}
		logger.Debugf("[ Socket ] listen Udp :%s", udpAddr)

		co.Kcp = kcp
	}
	if co.Kcp == nil && co.Tcp == nil && co.WebSocket == nil {
		return fmt.Errorf("[ Socket ] webSocket、 TCP、 UDP , less select one")
	}
	co.stop = make(chan bool)
	co.conn = make(chan Conn)
	co.listenChannel()
	return nil
}

func (co *core) Accept() <-chan Conn {
	return co.conn
}

func (co *core) listenChannel() {
	connHandle := func(accept <-chan protocol.Conn) {
		for {
			select {
			case c := <-accept:
				co.conn <- newConn(c)
			case <-co.stop:
				break
			}
			co.Close()
		}
	}
	if co.WebSocket != nil && co.WebSocket.Connecting() {
		go connHandle(co.WebSocket.Accept())
	}
	if co.Tcp != nil && co.Tcp.Connecting() {
		go connHandle(co.Tcp.Accept())
	}
	if co.Kcp != nil && co.Kcp.Connecting() {
		go connHandle(co.Kcp.Accept())
	}
}

func (co *core) Close() error {
	item := 0
	if co.WebSocket != nil && co.WebSocket.Connecting() {
		item += 1
		err := co.WebSocket.Close()
		if err != nil {
			return err
		}
	}
	if co.Tcp != nil && co.Tcp.Connecting() {
		item += 1
		err := co.Tcp.Close()
		if err != nil {
			return err
		}
	}
	if co.Kcp != nil && co.Kcp.Connecting() {
		item += 1
		err := co.Kcp.Close()
		if err != nil {
			return err
		}
	}
	for i := 0; i < item; i++ {
		co.stop <- true
	}
	return nil
}

type conn struct {
	current   protocol.Conn
	allowConn []protocol.Conn

	idMx sync.Mutex
	id   int
	cm   map[int][]interface{}
}

func newConn(pc protocol.Conn) Conn {
	c := conn{}
	c.current = pc

}

func (c *conn) listenChannel() {

}

func (c *conn) Session() string {
	return ""
}

func (c *conn) Disconnect() {

}

func (c *conn) Write(b []byte) {

}

func (c *conn) Accept() (int, <-chan []byte, <-chan string, <-chan error) {
	c.idMx.Lock()
	id := c.id
	c.id = c.id + 1
	c.idMx.Unlock()
	d := make(chan []byte)
	cl := make(chan string)
	e := make(chan error)
	c.cm[id] = []interface{}{d, cl, e}

	return id, d, cl, e
}

func (c *conn) UnAccept(id int) {
	if id >= c.id {
		return
	}
	l, ok := c.cm[id]
	if !ok {
		return
	}
	for _, cc := range l {
		close(cc.(chan interface{}))
	}
	delete(c.cm, id)
}
