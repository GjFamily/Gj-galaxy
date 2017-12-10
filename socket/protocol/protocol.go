package protocol

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"net"
)

type Protocol interface {
	Accept() (<-chan net.Conn, error)
	Close() error
	Connecting() bool
}

type OptimizeProtocol interface {
	Protocol
	Optimize(conn net.Conn, delay int64)
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

func intactOut(b []byte) []byte {
	packetLength := len(b) + 8
	result := make([]byte, packetLength)
	result[0] = 0xFF
	result[1] = 0xFF
	result[2] = byte(uint16(len(b)) >> 8)
	result[3] = byte(uint16(len(b)) & 0xFF)
	copy(result[4:], b)
	sendCrc := crc32.ChecksumIEEE(b)
	result[packetLength-4] = byte(sendCrc >> 24)
	result[packetLength-3] = byte(sendCrc >> 16 & 0xFF)
	result[packetLength-2] = 0xFF
	result[packetLength-1] = 0xFE
	return result
}

func intactIn(b []byte) []byte {

}
