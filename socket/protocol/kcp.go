package protocol

import (
	"fmt"
	"net"

	kcpGo "github.com/xtaci/kcp-go"
)

type kcp struct {
	listener   *kcpGo.Listener
	connecting bool
	accept     chan net.Conn
}

func KcpListen(udpAddr *net.UDPAddr) (OptimizeProtocol, error) {
	k := kcp{}
	listener, err := kcpGo.ListenWithOptions(udpAddr.String(), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	k.listener = listener
	k.connecting = true
	return &k, nil
}

func (socket *kcp) Accept() (<-chan net.Conn, error) {
	if socket.accept != nil {
		return socket.accept, fmt.Errorf("accept already been use")
	}
	socket.accept = make(chan net.Conn)
	go func() {
		defer socket.Close()
		for {
			if !socket.connecting {
				break
			}
			conn, err := socket.listener.AcceptKCP()
			if err != nil {

			} else {
				//https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
				conn.SetStreamMode(true)
				conn.SetWriteDelay(true)
				conn.SetNoDelay(0, 20, 2, 1)
				conn.SetMtu(1024)
				conn.SetWindowSize(1024, 1024)
				conn.SetACKNoDelay(false)
				socket.accept <- conn
			}
		}
		close(socket.accept)
	}()
	return socket.accept, nil
}

func (socket *kcp) Close() error {
	if !socket.connecting {
		return nil
	}
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

func (socket *kcp) Optimize(conn net.Conn, delay int64) {
	// 根据延迟情况调整
}
