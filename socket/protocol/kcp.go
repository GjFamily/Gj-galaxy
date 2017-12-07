package protocol

import (
	"net"

	kcpGo "github.com/xtaci/kcp-go"
)

type kcp struct {
	listener   *kcpGo.Listener
	connecting bool
	accept     chan Conn
}

func KcpListen(udpAddr *net.UDPAddr) (Protocol, error) {
	k := kcp{}
	listener, err := kcpGo.ListenWithOptions(udpAddr.String(), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	k.listener = listener
	k.connecting = true
	k.accept = make(chan Conn)
	return &k, nil
}

func (socket *kcp) Accept() <-chan Conn {
	return socket.accept
}

func (socket *kcp) Close() error {
	err := socket.listener.Close()
	if err != nil {
		return err
	}
	socket.connecting = false
	return nil
}

func (socket *kcp) Connecting() bool {
	return socket.connecting
}
