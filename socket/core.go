package socket

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"sync/atomic"
	"time"

	"Gj-galaxy/socket/protocol"
)

type ProtocolType string

const (
	DefaultProtocol ProtocolType = "tcp"
	SpeedProtocol   ProtocolType = "udp"
	SafeProtocol    ProtocolType = "websocket"
)

type Core interface {
	Listen(tcpAddr *net.TCPAddr, udpAddr *net.UDPAddr) error
	Attach(path string, http HttpMux) error
	Close() error
}

type core struct {
	WebSocket protocol.Protocol
	Tcp       protocol.Protocol
	Udp       protocol.OptimizeProtocol

	e           *engine
	Clients     map[string]*client
	ClientCount int64
}

func newCore(e *engine) Core {
	co := core{}
	co.e = e
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
		co.e.Logger.Debugf("[ Socket ] listen Tcp :%s", tcpAddr)
		co.Tcp = tcp
	}
	if udpAddr != nil {
		udp, err := protocol.KcpListen(udpAddr)
		if err != nil {
			return err
		}
		co.e.Logger.Debugf("[ Socket ] listen Udp :%s", udpAddr)

		co.Udp = udp
	}
	if co.Udp == nil && co.Tcp == nil && co.WebSocket == nil {
		return fmt.Errorf("[ Socket ] webSocket、 TCP、 UDP , less select one")
	}
	co.listenChannel()
	return nil
}

func (co *core) listenChannel() {
	connHandle := func(protocol protocol.Protocol, t ProtocolType) error {
		if protocol == nil || protocol.Connecting() {
			return fmt.Errorf("%s is not able", t)
		}
		accept, err := protocol.Accept()
		if err != nil {
			return err
		}
		go func() {
			defer protocol.Close()
			for c := range accept {
				co.newConn(t, c)
			}
		}()
		return nil
	}
	connHandle(co.WebSocket, SafeProtocol)
	connHandle(co.Tcp, DefaultProtocol)
	connHandle(co.Udp, SpeedProtocol)
}

func (co *core) newConn(t ProtocolType, conn net.Conn) {
	go func() {
		// close connection before exit
		defer conn.Close()
		// 状态机状态
		state := 0x00
		// 数据包长度
		length := uint16(0)
		// crc校验和
		crc16 := uint16(0)
		var recvBuffer []byte
		// 游标
		cursor := uint16(0)
		bufferReader := bufio.NewReader(conn)
		//状态机处理数据
		for {
			recvByte, err := bufferReader.ReadByte()
			if err != nil {
				//这里因为做了心跳，所以就没有加deadline时间，如果客户端断开连接
				//这里ReadByte方法返回一个io.EOF的错误，具体可考虑文档
				if err == io.EOF {
					fmt.Printf("client %s is close!\n", conn.RemoteAddr().String())
				}
				//在这里直接退出goroutine，关闭由defer操作完成
				return
			}
			//进入状态机，根据不同的状态来处理
			switch state {
			case 0x00:
				if recvByte == 0xFF {
					state = 0x01
					//初始化状态机
					recvBuffer = nil
					length = 0
					crc16 = 0
				} else {
					state = 0x00
				}
				break
			case 0x01:
				if recvByte == 0xFF {
					state = 0x02
				} else {
					state = 0x00
				}
				break
			case 0x02:
				length += uint16(recvByte) * 256
				state = 0x03
				break
			case 0x03:
				length += uint16(recvByte)
				// 一次申请缓存，初始化游标，准备读数据
				recvBuffer = make([]byte, length)
				cursor = 0
				state = 0x04
				break
			case 0x04:
				//不断地在这个状态下读数据，直到满足长度为止
				recvBuffer[cursor] = recvByte
				cursor++
				if cursor == length {
					state = 0x05
				}
				break
			case 0x05:
				crc16 += uint16(recvByte) * 256
				state = 0x06
				break
			case 0x06:
				crc16 += uint16(recvByte)
				state = 0x07
				break
			case 0x07:
				if recvByte == 0xFF {
					state = 0x08
				} else {
					state = 0x00
				}
			case 0x08:
				if recvByte == 0xFE {
					//执行数据包校验
					if (crc32.ChecksumIEEE(recvBuffer)>>16)&0xFFFF == uint32(crc16) {
						message := &Message{}
						message.Unpack(recvBuffer)
						//新开协程处理数据
						go co.dispatch(conn, t, message)
					} else {
						fmt.Println("丢弃数据!")
					}
				}
				//状态机归位,接收下一个包
				state = 0x00
			}
		}
	}()
}

func (co *core) dispatch(conn net.Conn, t ProtocolType, message *Message) {
	sid := message.SID
	if message.Type == P_OPEN {
		if sid == "" {
			sid = UniqueId()
			client := newClient(co.e, co, sid)
			co.Clients[sid] = client
			atomic.AddInt64(&co.ClientCount, 1)
			return
		}
	}
	client := co.Clients[sid]
	switch message.Type {
	case P_CLOSE:
		client.onClose()
		delete(co.Clients, sid)
		atomic.AddInt64(&co.ClientCount, -1)
	case P_PING:
		if t == SpeedProtocol {
			var tmp int64

			bytesBuffer := bytes.NewBuffer(message.Data)
			binary.Read(bytesBuffer, binary.BigEndian, &tmp)
			delay := tmp - time.Now().UnixNano()/1e6
			co.Udp.Optimize(conn, delay)
		}
		client.onPing(message, t)
	case P_MESSAGE:
		client.onMessage(message)
	case P_PROTOCOL:
		protocolName := ProtocolType(message.Data[:])
		client.onProtocol(t, protocolName)
	}
}
func (co *core) Close() error {
	closeHandle := func(protocol protocol.Protocol) error {
		if protocol != nil && protocol.Connecting() {
			err := protocol.Close()
			if err != nil {
				return err
			}
		}
		return nil
	}
	closeHandle(co.WebSocket)
	closeHandle(co.Tcp)
	closeHandle(co.Udp)

	for _, client := range co.Clients {
		client.Close()
	}
	co.Clients = nil
	co.ClientCount = 0
	return nil
}

func (co *core) Message(message *Message, conn net.Conn) {
	sendBytes := message.Packet()
	packetLength := len(sendBytes) + 8
	result := make([]byte, packetLength)
	result[0] = 0xFF
	result[1] = 0xFF
	result[2] = byte(uint16(len(sendBytes)) >> 8)
	result[3] = byte(uint16(len(sendBytes)) & 0xFF)
	copy(result[4:], sendBytes)
	sendCrc := crc32.ChecksumIEEE(sendBytes)
	result[packetLength-4] = byte(sendCrc >> 24)
	result[packetLength-3] = byte(sendCrc >> 16 & 0xFF)
	result[packetLength-2] = 0xFF
	result[packetLength-1] = 0xFE
	conn.Write(result)
}

func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
