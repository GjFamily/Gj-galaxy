package socket

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

type Client interface {
	Packet(packet *Data, opts *packetOpt)
	Connect(s *socketInline)
	Disconnect(s *socketInline)
	Close()
}

// 消息通道，链接状态
type client struct {
	sid           string
	e             *engine
	co            *core
	nsp           map[string]*socketInline
	connectBuffer []string
	stop          chan bool
	listening     bool

	allowConn map[ProtocolType]net.Conn
}

func newClient(e *engine, co *core, sid string) *client {
	c := client{
		sid,
		e,
		co,
		make(map[string]*socketInline),
		make([]string, 0),
		make(chan bool),
		false,
		make(map[ProtocolType]net.Conn),
	}
	c.listenChannel()
	return &c
}

func (c *client) Open(t ProtocolType, conn net.Conn) {
	if old, ok := c.allowConn[t]; ok {
		old.Close()
	}
	c.allowConn[t] = conn
}

func (c *client) listenChannel() {
	go func() {
		c.listening = true
		for {
			select {
			case <-c.stop:
				break
			}
		}
	}()
}

func (c *client) Message(message *Message, protocolType ProtocolType) {
	message.SID = c.sid
	conn, ok := c.allowConn[protocolType]
	if !ok {
		for protocolTmp, _conn := range c.allowConn {
			conn = _conn
			c.onProtocol(protocolTmp, protocolType)
			break
		}
	}
	c.co.Message(message, conn)
}

func (c *client) Packet(packet *Data, opts *packetOpt) {
	c.Message(&Message{Type: P_MESSAGE, Data: packet.Packet()}, opts.protocolType)
}

func (c *client) Close() {
	for _, s := range c.nsp {
		s.Disconnect()
	}
	c.destroy("forced server close")
}

func (c *client) Disconnect(s *socketInline) error {
	s, ok := c.nsp[s.namespace.name]
	if !ok {
		return fmt.Errorf("[ SOCKET ] no socket for namespace %s", s.namespace.name)
	}
	delete(c.nsp, s.namespace.name)

	return nil
}

func (c *client) onConnect(s *socketInline) {
	c.nsp[s.namespace.name] = s
}

func (c *client) valid(name string) {
	var nsp, ok = c.e.nss[name]
	if !ok {
		c.Packet(&Data{Type: P_ERROR, NSP: name, Data: "Invalid namespace"}, nil)
		return
	}
	nsp.clientChannel <- c
}

func (c *client) GetSession() string {
	return c.sid
}

func (c *client) onMessage(message *Message) {
	defer func() {
		if err := recover(); err != nil {
			c.onError(errors.New(err.(string)))
		}
	}()
	p := &Data{}
	p.Unpack(message.Data)
	c.onDecode(p)
}

func (c *client) onPing(message *Message, protocolType ProtocolType) {
	message.Type = P_PONG
	c.Message(message, protocolType)
}

func (c *client) onProtocol(protocolType ProtocolType, destType ProtocolType) {
	message := &Message{}
	// todo 控制发送协议数据的频率
	switch destType {
	case SpeedProtocol:
		message.Data = []byte(c.e.UDPAddr.String())
	case DefaultProtocol:
		message.Data = []byte(c.e.TCPAddr.String())
	}
	// 已连接的协议，会在下次链接时自动关闭
	c.Message(message, protocolType)
}

func (c *client) onDecode(packet *Data) {
	u, err := url.Parse(packet.NSP)
	if err != nil {
		c.destroy("namespace format error")
		return
	}
	if packet.Type == P_CONNECT {
		c.valid(u.Path)
	} else {
		s, ok := c.nsp[u.Path]
		if ok {
			switch packet.Type {
			case P_EVENT:
				s.onEvent(packet)
			case P_ACK:
				s.onAck(packet)
			case P_DISCONNECT:
				s.onDisconnect()
			case P_ERROR:
				s.onError(errors.New(packet.Data.(string)))
			}
		} else {
			c.e.Logger.Debugf("[ SOCKET ] no socket for namespace %s", packet.NSP)
		}
	}
}

func (c *client) onError(err error) {
	for _, s := range c.nsp {
		s.onError(err)
	}
	//c.destroy("client data error")
}

func (c *client) onClose() {
	for _, s := range c.nsp {
		s.onClose()
	}
	c.destroy("client close")
}

func (c *client) destroy(reason string) {
	for _, conn := range c.allowConn {
		conn.Close()
	}
	if c.listening {
		c.stop <- true
	}
}
