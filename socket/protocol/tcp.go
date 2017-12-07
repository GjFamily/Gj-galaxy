package protocol

import "net"

type tcp struct {
	listener   *net.TCPListener
	connecting bool
	accept     chan Conn
}

func TcpListen(tcpAddr *net.TCPAddr) (Protocol, error) {
	t := tcp{}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	t.listener = listener
	t.connecting = true
	t.accept = make(chan Conn)
	return &t, nil
}

func (socket *tcp) Accept() <-chan Conn {
	return socket.accept
}

func (socket *tcp) Close() error {
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

func (socket *tcp) Connecting() bool {
	return socket.connecting
}
