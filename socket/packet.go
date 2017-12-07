package socket

import (
	"bytes"
	"encoding/binary"
)

type packetType int

const (
	P_CONNECT packetType = iota
	P_DISCONNECT
	P_EVENT
	P_ACK
	P_ERROR
)

type packet struct {
	Type         packetType
	NSP          string
	Id           int
	Data         interface{}
	attachNumber int
}

type packetOpt struct {
	compress bool
}

//封包
func (p *packet) Packet() []byte {
	return []byte{}
}

//解包
func (p *packet) Unpack(buffer []byte) {

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
