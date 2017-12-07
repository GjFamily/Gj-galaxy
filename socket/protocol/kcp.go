package protocol

import (
	"net"

	kcpGo "github.com/xtaci/kcp-go"
)

type kcp struct {
	listener   *kcpGo.Listener
	connecting bool
}

func KcpListen(udp *net.UDPAddr) (Protocol, error) {
	k := kcp{}
	listener, err := kcpGo.ListenWithOptions(udp.String(), block, config.DataShard, config.ParityShard)
	if err != nil {
		return nil, err
	}
	k.listener = listener
	k.connecting = true
	return &k, nil
}

func (socket *kcp) Accept() <-chan Conn {

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
