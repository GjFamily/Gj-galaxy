package protocol

import (
	"net/http"

	"golang.org/x/net/websocket"
)

type Core interface {
}

type HttpMux interface {
	Handle(pattern string, handler http.Handler)
}
type webSocket struct {
	ws         *websocket.Conn
	connecting bool
	accept     chan Conn
}

func WebSocketAttach(path string, http HttpMux) (Protocol, error) {
	socket := webSocket{}

	handler := func(ws *websocket.Conn) {
		socket.connecting = true
		socket.ws = ws
	}
	http.Handle(path, websocket.Handler(handler))
	return &socket, nil
}

func (socket *webSocket) Accept() <-chan Conn {
	return socket.accept
}

func (socket *webSocket) Close() error {
	return nil
}

func (socket *webSocket) Connecting() bool {
	return socket.connecting
}

func WebSocket(ws *websocket.Conn) {
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
