package gipc

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func GetTimeout() time.Duration {
	envVar := os.Getenv("GIPC_TIMEOUT")
	if len(envVar) > 0 {
		valInt, err := strconv.Atoi(envVar)
		if err == nil {
			return time.Duration(valInt)
		}
	}
	return 3
}

var TimeoutClientConfig = ClientConfig{
	Timeout:    time.Second * GetTimeout(),
	Encryption: ENCRYPT_BY_DEFAULT,
}

func NewClientTimeoutConfig(name string) *ClientConfig {
	config := TimeoutClientConfig
	config.Name = name
	return &config
}

func NewServerConfig(name string) *ServerConfig {
	return &ServerConfig{Name: name, Encryption: ENCRYPT_BY_DEFAULT}
}

func NewClientConfig(name string) *ClientConfig {
	return &ClientConfig{Name: name, Encryption: ENCRYPT_BY_DEFAULT}
}

func TestBaseStartUp_Name(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig(""))
	if err.Error() != "ipcName cannot be an empty string" {
		t.Error("server - should have an error becuse the ipc name is empty")
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig(""))
	if err2.Error() != "ipcName cannot be an empty string" {
		t.Error("client - should have an error becuse the ipc name is empty")
	}
	defer cc.Close()
}

func TestBaseStartUp_Configs(t *testing.T) {

	Sleep()

	scon := NewServerConfig("test_config")
	ccon := NewClientConfig("test_config")

	sc, err3 := StartServer(scon)
	if err3 != nil {
		t.Error(err3)
	}
	defer sc.Close()

	Sleep()

	cc, err4 := StartClient(ccon)
	if err4 != nil {
		t.Error(err4)
	}
	defer cc.Close()
}

func TestBaseStartUp_Configs2(t *testing.T) {

	Sleep()

	scon := NewServerConfig("test_config2")
	ccon := NewClientConfig("test_config2")

	scon.MaxMsgSize = -1

	sc, err5 := StartServer(scon)
	if err5 != nil {
		t.Error(err5)
	}
	defer sc.Close()

	Sleep()

	//testing junk values that will default to 0
	ccon.Timeout = -1
	ccon.RetryTimer = -1

	cc, err6 := StartClient(ccon)
	if err6 != nil {
		t.Error(err6)
	}
	defer cc.Close()
}

func TestBaseStartUp_Configs3(t *testing.T) {

	Sleep()

	scon := NewServerConfig("test_config3")
	ccon := NewClientConfig("test_config3")

	scon.MaxMsgSize = 1025

	sc, err7 := StartServer(scon)
	if err7 != nil {
		t.Error(err7)
	}
	defer sc.Close()

	Sleep()

	cc, err8 := StartClient(ccon)
	if err8 != nil {
		t.Error(err8)
	}
	defer cc.Close()
}

func TestBaseTimeoutNoServer(t *testing.T) {

	Sleep()

	ccon := NewClientTimeoutConfig("test_timeout")
	cc, err := StartClient(ccon)
	defer cc.Close()

	if !strings.Contains(err.Error(), "timed out trying to connect") {
		t.Error(err)
	}
}

func TestBaseTimeoutNoServerRetry(t *testing.T) {

	Sleep()

	dialFinished := make(chan bool, 1)

	go func() {
		time.Sleep(time.Second * 4)
		dialFinished <- true
	}()

	go func() {
		//this should retry every second and never return
		cc, err := StartClient(NewClientTimeoutConfig("test_timeout_retryloop"))
		defer cc.Close()
		if err != nil && !strings.Contains(err.Error(), "timed out trying to connect") {
			t.Error(err)
		}
	}()

	<-dialFinished
}

func TestBaseTimeoutServerDisconnected(t *testing.T) {

	Sleep()

	scon := NewServerConfig("test_timeout_server_disconnect")
	sc, err := StartServer(scon)
	if err != nil {
		t.Error(err)
	}

	ccon := NewClientTimeoutConfig("test_timeout_server_disconnect")

	Sleep()

	cc, err2 := StartClient(ccon)
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	go func() {
		time.Sleep(time.Second * 1)
		sc.Close()
	}()

	for {
		_, err := cc.Read() //Timed(time.Second*2, TimeoutMessage)
		if err != nil {
			//this error will only be reached if Timeout is specified, otherwise
			//the reconnect dial loop will loop perpetually
			if err.Error() == "the received channel has been closed" {
				break
				//}
				//this will be the first error captured before received channel closure
			} else if err.Error() == "timed out trying to re-connect" {
				break
			}
		}
	}
}

func TestBaseRead(t *testing.T) {

	Sleep()

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
func TestBaseWrite(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test_write"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test_write"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)

	go func() {

		for {

			m, err := cc.Read()
			if err != nil {
				//t.Error(fmt.Sprintf("Got read error: %s", err))
			} else if m.Status == "Connected" {
				connected <- true
				return
			}
		}
	}()

	<-connected

	buf := make([]byte, 1)

	err3 := sc.Write(0, buf)

	if err3.Error() != "message type 0 is reserved" {
		t.Error("0 is not allowed as a message type")
	}

	buf = make([]byte, sc.Config.ServerConfig.MaxMsgSize+5)
	err4 := sc.Write(2, buf)

	if err4.Error() != "message exceeds maximum message length" {
		t.Errorf("There should be an error as the data we're attempting to write is bigger than the MAX_MSG_SIZE, instead we got: %s", err4)
	}

	sc.setStatus(NotConnected)

	buf2 := make([]byte, 5)
	err5 := sc.Write(2, buf2)
	if err5.Error() != "cannot write under current status: Not Connected" {
		t.Errorf("we should have an error becuse there is no connection but instead we got: %s", err5)
	}

	sc.setStatus(Connected)

	buf = make([]byte, 1)

	err = cc.Write(0, buf)
	if err == nil {
		t.Error("0 is not allowwed as a message try")
	}

	buf = make([]byte, MAX_MSG_SIZE+5)
	err = cc.Write(2, buf)
	if err == nil {
		t.Error("There should be an error is the data we're attempting to write is bigger than the MAX_MSG_SIZE")
	}

	cc.setStatus(NotConnected)

	buf = make([]byte, 5)
	err = cc.Write(2, buf)
	if err.Error() == "cannot write under current status: Not Connected" {

	} else {
		t.Error("we should have an error becuse there is no connection")
	}
}

func TestBaseStatus(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test_status"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	sc.setStatus(NotConnected)

	if sc.Status() != "Not Connected" {
		t.Error("status string should have returned Not Connected")
	}

	sc.setStatus(Listening)

	if sc.Status() != "Listening" {
		t.Error("status string should have returned Listening")
	}

	sc.setStatus(Connecting)

	if sc.Status() != "Connecting" {
		t.Error("status string should have returned Connecting")
	}

	sc.setStatus(Connected)

	if sc.Status() != "Connected" {
		t.Error("status string should have returned Connected")
	}

	sc.setStatus(ReConnecting)

	if sc.Status() != "Reconnecting" {
		t.Error("status string should have returned Reconnecting")
	}

	sc.setStatus(Closed)

	if sc.Status() != "Closed" {
		t.Error("status string should have returned Closed")
	}

	sc.setStatus(Error)

	if sc.Status() != "Error" {
		t.Error("status string should have returned Error")
	}

	sc.setStatus(Closing)

	if sc.Status() != "Closing" {
		t.Error("status string should have returned Error")
	}
}

func TestBaseGetConnected(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test22"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test22"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	for {
		cc.Read()
		m, _ := sc.Read()

		if m.Status == "Connected" {
			break
		}
	}
}

func TestBaseServerWrongMessageType(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test333"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test333"))
	if err2 != nil {
		t.Error(err)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {

		ready := false

		for {
			m, _ := sc.Read()
			if m.Status == "Connected" {
				connected <- true
				ready = true
				continue
			}

			if ready == true {
				if m.MsgType != 5 {
					// received wrong message type

				} else {
					t.Error("should have got wrong message type")
				}
				complete <- true
				break
			}
		}

	}()

	go func() {
		for {
			m, _ := cc.Read()

			if m.Status == "Connected" {
				connected2 <- true
				return
			}
		}
	}()

	<-connected
	<-connected2

	// test wrong message type
	cc.Write(2, []byte("hello server 1"))

	<-complete
}
func TestBaseClientWrongMessageType(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test3"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test3"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {
		for {
			m, _ := sc.Read()
			if m.Status == "Connected" {
				connected2 <- true
				return
			}

		}
	}()

	go func() {

		ready := false

		for {

			m, err45 := cc.Read()

			if m.Status == "Connected" {
				connected <- true
				ready = true
				continue

			}

			if ready == true {

				if err45 == nil {
					if m.MsgType != 5 {
						// received wrong message type
					} else {
						t.Error("should have got wrong message type")
					}
					complete <- true
					break

				} else {
					t.Error(err45)
					break
				}
			}

		}
	}()

	<-connected
	<-connected2
	sc.Write(2, []byte(""))

	<-complete
}
func TestBaseServerCorrectMessageType(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test358"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test358"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {
		for {
			m, err := sc.Read()
			if err == nil && m.Status == "Connected" {
				connected2 <- true
				return
			}
		}
	}()

	go func() {

		ready := false

		for {
			m, err23 := cc.Read()
			if err23 == nil && m.Status == "Connected" {
				ready = true
				connected <- true
				continue
			}
			if ready == true {
				if err23 == nil {
					if m.MsgType == 5 {
						// received correct message type
					} else {
						t.Error("should have got correct message type")
					}

					complete <- true
					return
				} else {
					t.Error(err23)
					break
				}
			}

		}
	}()

	<-connected
	<-connected2

	sc.Write(5, []byte(""))

	<-complete
}

func TestBaseClientCorrectMessageType(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test355"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test355"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {

		for {
			m, _ := cc.Read()

			if m.Status == "Connected" {
				connected2 <- true
				return
			}
		}

	}()

	go func() {

		ready := false

		for {

			m, err34 := sc.Read()

			if m.Status == "Connected" {
				ready = true
				connected <- true
				continue
			}

			if ready == true {
				if err34 == nil {
					if m.MsgType == 5 {
						// received correct message type
					} else {
						t.Error("should have got correct message type")
					}

					complete <- true
					return
				} else {
					t.Error(err34)
					break
				}
			}
		}
	}()

	<-connected2
	<-connected

	cc.Write(5, []byte(""))
	<-complete
}

func TestBaseServerSendMessage(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test377"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test377"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {

		for {

			m, err := sc.Read()
			if err != nil {
				t.Error(fmt.Sprintf("Got read error: %s", err))
			} else if m.Status == "Connected" {
				connected <- true
				return
			}
		}
	}()

	go func() {

		ready := false

		for {

			m, err56 := cc.Read()

			if m.Status == "Connected" {
				ready = true
				connected2 <- true
				continue
			}

			if ready == true {
				if err56 == nil {
					if m.MsgType == 5 {
						if string(m.Data) == "Here is a test message sent from the server to the client... -/and some more test data to pad it out a bit" {
							// correct msg has been received
						} else {
							t.Error("Message recreceivedieved is wrong")
						}
					} else {
						t.Error("should have got correct message type")
					}

					complete <- true
					break

				} else {
					t.Error(err56)
					complete <- true
					break
				}

			}
		}

	}()

	<-connected2
	<-connected

	sc.Write(5, []byte("Here is a test message sent from the server to the client... -/and some more test data to pad it out a bit"))

	<-complete
}
func TestBaseClientSendMessage(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test3661"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test3661"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	connected2 := make(chan bool, 1)
	complete := make(chan bool, 1)

	go func() {

		for {

			m, _ := cc.Read()
			if m.Status == "Connected" {
				connected <- true
				return
			}

		}
	}()

	go func() {

		ready := false

		for {

			m, _ := sc.Read()

			if m.Status == "Connected" {
				ready = true
				connected2 <- true
				continue
			}

			if ready == true {
				if err == nil {
					if m.MsgType == 5 {

						if string(m.Data) == "Here is a test message sent from the client to the server... -/and some more test data to pad it out a bit" {
							// correct msg has been received
						} else {
							t.Error("Message recreceivedieved is wrong")
						}

					} else {
						t.Error("should have got correct message type")
					}
					complete <- true
					break

				} else {
					t.Error(err)
					complete <- true
					break
				}
			}

		}
	}()

	<-connected
	<-connected2

	cc.Write(5, []byte("Here is a test message sent from the client to the server... -/and some more test data to pad it out a bit"))

	<-complete
}

func TestBaseClientClose(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test10A"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()

	cc, err2 := StartClient(NewClientConfig("test10A"))
	if err2 != nil {
		t.Error(err2)
	}

	holdIt := make(chan bool, 1)

	go func() {

		for {

			m, _ := sc.Read()

			if m.Status == "Disconnected" {
				holdIt <- false
				break
			}

		}

	}()

	for {

		mm, err := cc.Read()

		if err == nil {
			if mm.Status == "Connected" {
				cc.Close()
			}

			if mm.Status == "Closed" {
				break
			}
		}

	}

	<-holdIt
}

func TestBaseClientReadClose(t *testing.T) {

	Sleep()

	sc, err := StartServer(NewServerConfig("test_clientReadClose"))
	if err != nil {
		t.Error(err)
	}

	Sleep()

	cc, err2 := StartClient(NewClientTimeoutConfig("test_clientReadClose"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	connected := make(chan bool, 1)
	clientTimout := make(chan bool, 1)
	clientConnected := make(chan bool, 1)
	clientError := make(chan bool, 1)

	go func() {

		for {

			m, _ := sc.Read()
			if m.Status == "Connected" {
				connected <- true
				break
			}

		}
	}()

	go func() {

		reconnect := false

		for {

			m, err3 := cc.Read()

			if err3 != nil {
				log.Printf("err: %s", err3)
				if err3.Error() == "the received channel has been closed" {
					clientError <- true // after the connection times out the received channel is closed, so we're now testing that the close error is returned.
					// This is the only error the received function returns.
					break
				}
			}

			if err3 == nil {
				if m.Status == "Connected" {
					clientConnected <- true
				} else if m.Status == "Reconnecting" {
					reconnect = true
				} else if m.Status == "Timeout" && reconnect == true {
					clientTimout <- true
				}
			}
		}
	}()

	<-connected
	<-clientConnected
	//IMPORTANT Close was not placed here by mistake
	sc.Close()
	<-clientTimout
	<-clientError
}

func TestBaseServerWrongEncryption(t *testing.T) {

	Sleep()

	scon := NewServerConfig("testl337_enc")
	scon.Encryption = false
	sc, err := StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()
	ccon := NewClientConfig("testl337_enc")
	ccon.Encryption = true
	cc, err2 := StartClient(ccon)
	defer cc.Close()

	if err2 != nil {
		if err2.Error() != "server tried to connect without encryption" {
			t.Error(err2)
		}
	}

	go func() {
		for {
			m, err := cc.Read()
			cc.logger.Debugf("Message: %v, err %s", m, err)
			if err != nil {
				if err.Error() != "server tried to connect without encryption" && m.MsgType != -2 {
					t.Error(err)
				}
				break
			} else if m.Status == "Closed" {
				break
			}
		}
	}()

	for {
		mm, err2 := sc.Read()
		sc.logger.Debugf("Message: %v, err %s", mm, err)
		if err2 != nil {
			if err2.Error() != "client is enforcing encryption" && mm.MsgType != -2 {
				t.Error(err2)
			}
			break
		}
	}
}

func TestBaseServerWrongEncryption2(t *testing.T) {

	Sleep()

	scon := NewServerConfig("testl338_enc")
	scon.Encryption = true
	sc, err := StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	Sleep()
	ccon := NewClientConfig("testl338_enc")
	ccon.Encryption = false
	cc, err2 := StartClient(ccon)
	defer cc.Close()
	if err2 != nil {
		if err2.Error() != "server tried to connect without encryption" {
			t.Error(err2)
		}
	}

	go func() {
		for {
			m, err := cc.Read()
			cc.logger.Debugf("Message: %v, err %s", m, err)
			if err != nil {
				if err.Error() != "server tried to connect without encryption" {
					if m != nil && m.MsgType != -2 {
						t.Error(err)
					}
				}
				break
			} else if m.Status == "Closed" {
				break
			}
		}
	}()

	for {
		mm, err2 := sc.Read()
		sc.logger.Debugf("Message: %v, err %s", mm, err2)
		if err2 != nil {
			if err2.Error() != "public key received isn't valid length 97, got: 1" {
				t.Error(err2)
			}
			break
		}
	}
}
