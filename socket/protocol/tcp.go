package protocol

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

type tcp struct {
	listener   net.Listener
	connecting bool
	accept     chan net.Conn
}

func TcpListen(tcpAddr *net.TCPAddr) (Protocol, error) {
	t := tcp{}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	t.listener = listener
	t.connecting = true
	return &t, nil
}

func TslListen(tcpAddr *net.TCPAddr, certFile string, keyFile string) (Protocol, error) {
	crt, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{crt}
	// Time returns the current time as the number of seconds since the epoch.
	// If Time is nil, TLS uses time.Now.
	tlsConfig.Time = time.Now
	// Rand provides the source of entropy for nonces and RSA blinding.
	// If Rand is nil, TLS uses the cryptographic random reader in package
	// crypto/rand.
	// The Reader must be safe for use by multiple goroutines.
	tlsConfig.Rand = rand.Reader
	t := tcp{}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	t.listener = tls.NewListener(listener, tlsConfig)
	t.connecting = true
	return &t, nil
}

func (socket *tcp) Accept() (<-chan net.Conn, error) {
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
			conn, err := socket.listener.Accept()
			if err != nil {

			} else {
				socket.accept <- conn
			}
		}
		close(socket.accept)
	}()
	return socket.accept, nil
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

func (socket *tcp) tsl(certFile string, keyFile string) error {
	crt, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{crt}
	// Time returns the current time as the number of seconds since the epoch.
	// If Time is nil, TLS uses time.Now.
	tlsConfig.Time = time.Now
	// Rand provides the source of entropy for nonces and RSA blinding.
	// If Rand is nil, TLS uses the cryptographic random reader in package
	// crypto/rand.
	// The Reader must be safe for use by multiple goroutines.
	tlsConfig.Rand = rand.Reader

	socket.connecting = false
	socket.listener = tls.NewListener(socket.listener, tlsConfig)
	socket.connecting = true
	return nil
}
