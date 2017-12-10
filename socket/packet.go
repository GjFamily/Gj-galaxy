package socket

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

type DataType uint8

const (
	P_CONNECT DataType = iota
	P_DISCONNECT
	P_EVENT
	P_ACK
	P_ERROR
)

type Data struct {
	Type DataType    `json:"type"`
	Id   int         `json:"id"`
	NSP  string      `json:"nsp"`
	Data interface{} `json:"data"`
}

type packetOpt struct {
	compress     bool
	protocolType ProtocolType
}

//封包
func (p *Data) Packet() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic("data to json error")
	}
	return b
}

//解包
func (p *Data) Unpack(buffer []byte) {

}

type MessageType uint8

const (
	P_OPEN MessageType = iota
	P_CLOSE
	P_PING
	P_PONG
	P_PROTOCOL
	P_MESSAGE
)

type Message struct {
	Type MessageType
	SID  string
	Data []byte
}

//封包
func (m *Message) Packet() []byte {
	l := len(m.SID)
	result := make([]byte, 1+1+l+len(m.Data))
	result[0] = byte(m.Type & 0xFF)
	result[1] = byte(uint8(l) & 0XFF)
	copy(result[2:], m.SID)
	copy(result[len(m.SID)+1:], m.Data)
	return result
}

//解包
func (m *Message) Unpack(buffer []byte) {
	m.Type = MessageType(buffer[0])
	l := uint8(buffer[1])
	m.SID = string(buffer[2 : l+2])
	m.Data = buffer[l+2:]
}

//整形转换成字节
func IntToBytes(n int) []byte {
	x := int32(n)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}
