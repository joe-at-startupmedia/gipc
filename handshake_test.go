package gipc

import (
	"testing"
)

func serverConfig(name string) *ServerConfig {
	return &ServerConfig{Name: name, Encryption: ENCRYPT_BY_DEFAULT}
}

// client.connect is an internal method
func TestServerReceiveWrongVersionNumber(t *testing.T) {

	sc, err := StartServer(serverConfig("test_wrongversion"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	go func() {

		cc, err2 := NewClient("test_wrongversion", nil)
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
			if err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}

			return
		}

		if m.Err != nil {
			if m.Err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}

			return
		}
	}
}

// client.connect is an internal method
func TestServerReceiveWrongVersionNumberMulti(t *testing.T) {

	config := serverConfig("test_wrongversion_multi")
	config.MultiClient = true
	sc, err := StartServer(config)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	go func() {

		cc, err2 := NewClient("test_wrongversion_multi", &ClientConfig{MultiClient: true})
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
			if err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}

			return
		}

		if m.Err != nil {
			if m.Err.Error() != "client has a different VERSION number" {
				t.Error("should have error because server sent the client the wrong VERSION number 1")
			}

			return
		}
	}
}
