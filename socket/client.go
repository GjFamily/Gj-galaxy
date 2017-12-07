package socket

import (
	"errors"
	"fmt"
	"net/url"
)

type Client interface {
	Packet(packet *packet, opts *packetOpt)
}

// 消息通道，链接状态
type client struct {
	SID           string
	e             *engine
	sockets       []*socketInline
	nsp           map[string]*socketInline
	connectBuffer []string
	conn          Conn
	cid           int
	dataChan      <-chan []byte
	closeChan     <-chan string
	errorChan     <-chan error
	stop          chan bool
	listening     bool
}

func newClient(e *engine, conn Conn) Client {
	c := client{
		conn.Session(),
		e,
		make([]*socketInline, 1),
		make(map[string]*socketInline),
		make([]string, 1),
		conn,
		0,
		nil,
		nil,
		nil,

		make(chan bool),
		false,
	}
	c.listenChannel()
	return &c
}

func (c *client) listenChannel() {
	c.cid, c.dataChan, c.closeChan, c.errorChan = c.conn.Accept()
	go func() {
		c.listening = true
		for {
			select {
			case data := <-c.dataChan:
				c.onData(data)
			case clo := <-c.closeChan:
				c.onClose(clo)
			case err := <-c.errorChan:
				c.onError(err)
			case <-c.stop:
				break
			}
		}
		c.listening = false
	}()
}

func (c *client) Packet(packet *packet, opts *packetOpt) {
	c.conn.Write(packet.Packet())
}

func (c *client) Close() {
	for _, s := range c.sockets {
		s.Disconnect()
	}
	c.sockets = nil
	c.onClose("forced server close")
}

func (c *client) Remove(s *socketInline) error {
	s, ok := c.nsp[s.namespace.name]
	if !ok {
		return fmt.Errorf("[ SOCKET ] no socket for namespace %s", s.namespace.name)
	}
	delete(c.nsp, s.namespace.name)

	return nil
}

func (c *client) Connect(name string, sid string) {
	var nsp, ok = c.e.nss[name]
	if !ok {
		c.Packet(&packet{Type: P_ERROR, NSP: name, Data: "Invalid namespace"}, nil)
		return
	}
	if name != "/" && c.nsp["/"] == nil {
		c.connectBuffer = append(c.connectBuffer, name)
		return
	}
	s := <-nsp.add(c)
	c.sockets = append(c.sockets, s)
	c.nsp[name] = s
	for _, n := range c.connectBuffer {
		c.Connect(n, sid)
	}
}

func (c *client) onData(data []byte) {
	defer func() {
		if err := recover(); err != nil {
			c.onError(errors.New(err.(string)))
		}
	}()
	p := &packet{}
	p.Unpack(data)
	c.onDecode(p)
}

func (c *client) onDecode(packet *packet) {
	u, err := url.Parse(packet.NSP)
	if err != nil {
		c.onClose("namespace format error")
	}
	if packet.Type == P_CONNECT {
		c.Connect(u.Path, u.User.String())
	} else {
		s, ok := c.nsp[u.Path]
		if ok {
			s.onPacket(packet)
		} else {
			logger.Debugf("[ SOCKET ] no socket for namespace %s", packet.NSP)
		}
	}
}

func (c *client) onError(err error) {
	for _, s := range c.sockets {
		s.onError(err)
	}
	c.onClose("client data error")
}

func (c *client) onClose(reason string) {
	for _, s := range c.sockets {
		s.onClose(reason)
	}
	c.destroy()
}

func (c *client) destroy() {
	c.conn.UnAccept(c.cid)
	if c.listening {
		c.stop <- true
	}
	c.conn.Disconnect()
}
