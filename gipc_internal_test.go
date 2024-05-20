package gipc

import (
	"fmt"
	"testing"
	"time"
)

func serverConfig(name string) *ServerConfig {
	return &ServerConfig{Name: name, Encryption: ENCRYPT_BY_DEFAULT}
}

func TestRead(t *testing.T) {

	sIPC := &Server{Actor: Actor{
		status:   NotConnected,
		received: make(chan *Message),
	}}

	sIPC.status = Connected

	serverFinished := make(chan bool, 1)

	go func(s *Server) {

		_, err := sIPC.Read()
		if err != nil {
			t.Error("err should be nill as tbe read function should read the 1st message added to received")
		}
		_, err2 := sIPC.Read()
		if err2 != nil {
			t.Error("err should be nill as tbe read function should read the 1st message added to received")
		}

		_, err3 := sIPC.Read()
		if err3 == nil {
			t.Error("we should get an error as the messages have been read and the channel closed")

		} else {
			serverFinished <- true
		}

	}(sIPC)

	sIPC.received <- &Message{MsgType: 1, Data: []byte("message 1")}
	sIPC.received <- &Message{MsgType: 1, Data: []byte("message 2")}
	close(sIPC.received) // close channel

	<-serverFinished

	// Client - read tests

	// 3 x client side tests
	cIPC := &Client{
		Actor:      Actor{status: NotConnected, received: make(chan *Message)},
		timeout:    2 * time.Second,
		retryTimer: 1 * time.Second,
	}

	cIPC.status = Connected

	clientFinished := make(chan bool, 1)

	go func() {

		_, err4 := cIPC.Read()
		if err4 != nil {
			t.Error("err should be nill as tbe read function should read the 1st message added to received")
		}
		_, err5 := cIPC.Read()
		if err5 != nil {
			t.Error("err should be nill as tbe read function should read the 1st message added to received")
		}

		_, err6 := cIPC.Read()
		if err6 == nil {
			t.Error("we should get an error as the messages have been read and the channel closed")
		} else {
			clientFinished <- true
		}

	}()

	cIPC.received <- &Message{MsgType: 1, Data: []byte("message 1")}
	cIPC.received <- &Message{MsgType: 1, Data: []byte("message 1")}
	close(cIPC.received) // close received channel

	<-clientFinished
}

// client.connect is an internal method
func TestServerReceiveWrongVersionNumber(t *testing.T) {

	sc, err := StartServer(serverConfig("test5"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	go func() {

		cc, err2 := NewClient("test5", nil)
		if err2 != nil {
			t.Error(err2)
		}
		defer cc.Close()

		Sleep()
		//cc.ClientId = 1
		conn, err := cc.connect()
		if err != nil {
			t.Error(err)
		}
		cc.conn = conn

		recv := make([]byte, 2)
		_, err2 = cc.conn.Read(recv)
		if err2 != nil {
			return
		}

		if recv[0] != 4 {
			cc.handshakeSendReply(1)
			return
		}
	}()

	for {

		m, err := sc.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if m.Err != nil {
			if m.Err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}
		}
	}
}

// client.connect is an internal method
func TestServerReceiveWrongVersionNumberMulti(t *testing.T) {

	config := serverConfig("test5")
	config.MultiClient = true
	sc, err := StartServer(config)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	go func() {

		cc, err2 := NewClient("test5", &ClientConfig{MultiClient: true})
		if err2 != nil {
			t.Error(err2)
		}
		defer cc.Close()

		Sleep()
		cc.ClientId = 1
		conn, err := cc.connect()
		if err != nil {
			t.Error(err)
		}
		cc.conn = conn
		Sleep()
		recv := make([]byte, 2)
		_, err2 = conn.Read(recv)
		if err2 != nil {
			return
		}

		if recv[0] != 4 {
			cc.handshakeSendReply(1)
			return
		}
	}()

	for {

		m, err := sc.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if m.Err != nil {
			if m.Err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}
		}
	}
}
