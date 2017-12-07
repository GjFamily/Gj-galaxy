package protocol

import "net"

type tcp struct {
	listener   *net.TCPListener
	connecting bool
}

func TcpListen(tcp *net.TCPAddr) (Protocol, error) {
	t := tcp{}
	listener, err := net.ListenTCP("tcp", tcp)
	if err != nil {
		return nil, err
	}
	t.listener = listener
	t.connecting = true
	return &t, nil
}

func (socket *tcp) Accept() <-chan Conn {

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
