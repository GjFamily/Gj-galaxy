package protocol

import (
	"fmt"
	"net"
	"net/http"

	"github.com/golang/net/websocket"
)

type Core interface {
}

type HttpMux interface {
	Handle(pattern string, handler http.Handler)
}
type webSocket struct {
	connecting bool
	accept     chan net.Conn
}

func WebSocketAttach(path string, http HttpMux) (Protocol, error) {
	socket := webSocket{}

	handler := func(ws *websocket.Conn) {
		socket.connecting = true
		if socket.accept == nil {
			return
		}
		socket.accept <- ws
	}
	http.Handle(path, websocket.Handler(handler))
	return &socket, nil
}

func (socket *webSocket) Accept() (<-chan net.Conn, error) {
	if socket.accept != nil {
		return socket.accept, fmt.Errorf("accept already been use")
	}
	socket.accept = make(chan net.Conn)
	return socket.accept, nil
}

func (socket *webSocket) Close() error {
	socket.accept = nil
	return nil
}

func (socket *webSocket) Connecting() bool {
	return socket.connecting
}

func webSocketHandle(ws *websocket.Conn) {
	//var err error
	//var clientMessage string
	// use []byte if websocket binary type is blob or arraybuffer
	// var clientMessage []byte

	// cleanup on server side
	//defer func() {
	//	if err = ws.Close(); err != nil {
	//		log.Println("Websocket could not be closed", err.Error())
	//	}
	//}()
	//
	//client := ws.Request().RemoteAddr
	//log.Println("Client connected:", client)
	//sockCli := ClientConn{ws, client}
	//ActiveClients[sockCli] = ""
	//log.Println("Number of clients connected:", len(ActiveClients))
	//
	//// for loop so the websocket stays open otherwise
	//// it'll close after one Receieve and Send
	//for {
	//	if err = Message.Receive(ws, &clientMessage); err != nil {
	//		// If we cannot Read then the connection is closed
	//		log.Println("Websocket Disconnected waiting", err.Error())
	//		// remove the ws client conn from our active clients
	//		delete(ActiveClients, sockCli)
	//		log.Println("Number of clients still connected:", len(ActiveClients))
	//		return
	//	}
	//
	//	var msg_arr = strings.Split(clientMessage, "|")
	//	if msg_arr[0] == "login" && len(msg_arr) == 3 {
	//		name := msg_arr[1]
	//		pass := msg_arr[2]
	//
	//		if pass == User[name] {
	//			ActiveClients[sockCli] = name
	//
	//			if err = Message.Send(ws, "login|"+name); err != nil {
	//				log.Println("Could not send message to ", client, err.Error())
	//			}
	//		} else {
	//			log.Println("login faild:", clientMessage)
	//		}
	//
	//	} else if msg_arr[0] == "msg" {
	//		if ActiveClients[sockCli] != "" {
	//			clientMessage = "msg|" + time.Now().Format("2006-01-02 15:04:05") + " " + ActiveClients[sockCli] + " Said: " + msg_arr[1]
	//			for cs, na := range ActiveClients {
	//				if na != "" {
	//					if err = Message.Send(cs.websocket, clientMessage); err != nil {
	//						log.Println("Could not send message to ", cs.clientIP, err.Error())
	//					}
	//				}
	//			}
	//		}
	//	}
	//}
}
