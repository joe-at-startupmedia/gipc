package test

import (
	"github.com/joe-at-startupmedia/gipc"
	"testing"
	"time"
)

func TestReconnectClient(t *testing.T) {

	scon := NewServerConfig("test127")
	sc, err := gipc.StartServer(scon)
	if err != nil {
		t.Error(err)
	}

	gipc.Sleep()

	ccon := NewClientConfig("test127")

	cc, err2 := gipc.StartClient(ccon)
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()
	connected := make(chan bool, 1)
	clientConfirm := make(chan bool, 1)
	clientConnected := make(chan bool, 1)

	go func() {
		for {
			m, _ := sc.Read()
			if m.Status == "Connected" {
				connected <- true
				return
			}
		}
	}()

	go func() {

		reconnectCheck := false

		for {
			m, _ := cc.Read()
			if m == gipc.TimeoutMessage {
				continue
			} else if m != nil {
				if m.Status == "Connected" {
					if !reconnectCheck {
						clientConnected <- true
					} else {
						clientConfirm <- true
						return
					}
				} else if m.Status == "Reconnecting" {
					reconnectCheck = true
				}
			}
		}
	}()

	//wait for both the client and server to connect
	<-connected
	<-clientConnected

	//disconnect from the server
	sc.Close()

	//start a new server
	sc2, err := gipc.StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc2.Close()

	for {

		m, _ := sc2.Read()
		if m.Status == "Connected" {
			<-clientConfirm
			break
		}
	}
}

func TestReconnectClientTimeout(t *testing.T) {

	sc, err := gipc.StartServer(NewServerConfig("test7"))
	if err != nil {
		t.Error(err)
	}

	gipc.Sleep()

	cc, err2 := gipc.StartClient(NewClientTimeoutConfig("test7"))
	if err2 != nil {
		t.Error(err2)
	}
	defer cc.Close()

	go func() {

		for {

			m, _ := sc.Read()
			if m.Status == "Connected" {
				sc.Close()
				break
			}

		}
	}()

	connect := false
	reconnect := false

	for {

		mm, err5 := cc.Read()

		if err5 == nil {
			if mm.Status == "Connected" {
				connect = true
			} else if mm.Status == "Reconnecting" {
				reconnect = true
			} else if mm.Status == "Timeout" && reconnect == true && connect == true {
				return
			}
		} else {
			if err5.Error() != "timed out trying to re-connect" {
				t.Fatal("should have got the timed out error")
			}
			break
		}
	}
}

func TestReconnectServer(t *testing.T) {

	sc, err := gipc.StartServer(NewServerConfig("test1277"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	gipc.Sleep()

	ccon := NewClientConfig("test1277")
	cc, err2 := gipc.StartClient(ccon)
	if err2 != nil {
		t.Error(err2)
	}
	connected := make(chan bool, 1)
	clientConfirm := make(chan bool, 1)
	clientConnected := make(chan bool, 1)

	go func() {

		for {
			m, _ := cc.Read()
			if m != nil && m.Status == "Connected" {
				<-clientConnected
				connected <- true
			}
		}
	}()

	go func() {

		reconnectCheck := 0
		connectCnt := 0
		for {
			m, err := sc.Read()

			if err != nil {
				//sc.logger.Debugf("TestServerReconnect sever read loop err: %s", err)
				//t.Error(err)
				return
			}

			if m.Status == "Connected" {
				if reconnectCheck == 1 && connectCnt > 0 {
					clientConfirm <- true
				} else {
					clientConnected <- true
					connectCnt++
				}
			}

			//dispatched on EOF
			if m.Status == "Disconnected" {
				reconnectCheck = 1
			}

		}
	}()

	<-connected
	cc.Close()

	//THIS IS IMPORTANT
	// this allows time for the server to realize the client disconnected
	// before adding the new client. If this is absent, tests will continue
	// working most of the time except when the rare race conditions are met
	time.Sleep(time.Second * 1)

	cc2, err := gipc.StartClient(NewClientTimeoutConfig("test1277"))
	if err != nil {
		t.Error(err)
	}
	defer cc2.Close()

	<-clientConfirm
}

func TestServerReconnect2(t *testing.T) {

	sc, err := gipc.StartServer(NewServerConfig("test337"))
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	gipc.Sleep()

	cc, err2 := gipc.StartClient(NewClientConfig("test337"))
	if err2 != nil {
		t.Error(err2)
	}

	hasConnected := make(chan bool, 1)
	hasDisconnected := make(chan bool, 1)
	hasReconnected := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-hasReconnected:
				return
			default:
				m, err := cc.Read()
				if m.Status == "Connected" {

					<-hasConnected

					cc.Close()

					<-hasDisconnected

					c2, err2 := gipc.StartClient(NewClientConfig("test337"))
					if err2 != nil {
						t.Error(err)
					}

					for {
						n, _ := c2.Read()
						if n.Status == "Connected" {
							c2.Close()
							break
						}
					}
					return
				}
			}
		}
	}()

	connect := false
	disconnect := false

	for {
		select {
		case <-hasReconnected:
			return
		default:
			m, _ := sc.Read()
			if m.Status == "Connected" && connect == false {
				hasConnected <- true
				connect = true
			}

			if m.Status == "Disconnected" {
				hasDisconnected <- true
				disconnect = true
			}

			if m.Status == "Connected" && connect == true && disconnect == true {
				hasReconnected <- true
			}
		}
	}
}

func TestReconnectServerMulti(t *testing.T) {

	scon := NewServerConfig("test1277_multi")
	scon.MultiClient = true
	sc, err := gipc.StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	gipc.Sleep()

	ccon := NewClientConfig("test1277_multi")
	ccon.MultiClient = true
	cc, err2 := gipc.StartClient(ccon)
	if err2 != nil {
		t.Error(err2)
	}

	connected := make(chan bool, 1)
	clientConfirm := make(chan bool, 1)
	clientConnected := make(chan bool, 1)

	go func() {

		for {
			m, _ := cc.Read()
			if m.Status == "Connected" {
				<-clientConnected
				connected <- true
				break
			}

		}
	}()

	go func() {

		reconnectCheck := 0

		for {
			//We used ReadTimed in order to scan client buffer on the next loop
			sc.Connections.ReadTimed(2*time.Second, func(_ *gipc.Server, m *gipc.Message, err error) {

				if err != nil {
					return
				}

				if m.Status == "Connected" {
					clientConnected <- true
				}

				if m.Status == "Disconnected" {
					reconnectCheck = 1
				}

				if m.Status == "Connected" && reconnectCheck == 1 {
					clientConfirm <- true
				}
			})
		}
	}()

	<-connected
	cc.Close()

	time.Sleep(1 * time.Second) //wait for connection to close before reconnecting
	ccon = NewClientConfig("test1277_multi")
	ccon.MultiClient = true
	c2, err := gipc.StartClient(ccon)
	if err != nil {
		t.Error(err)
	}
	defer c2.Close()

	for {
		m, _ := c2.Read()
		if m.Status == "Connected" {
			break
		}
	}

	<-clientConfirm
}

func TestReconnectServer2Multi(t *testing.T) {
	scon := NewServerConfig("test337_multi")
	scon.MultiClient = true
	sc, err := gipc.StartServer(scon)
	if err != nil {
		t.Error(err)
	}
	defer sc.Close()

	gipc.Sleep()
	ccon := NewClientConfig("test337_multi")
	ccon.MultiClient = true
	cc, err2 := gipc.StartClient(ccon)
	if err2 != nil {
		t.Error(err2)
	}

	hasConnected := make(chan bool, 1)
	hasDisconnected := make(chan bool, 1)
	hasReconnected := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-hasReconnected:
				return
			default:
				m, err := cc.Read()
				if m.Status == "Connected" {

					<-hasConnected

					cc.Close()

					<-hasDisconnected

					ccon2 := NewClientConfig("test337_multi")
					ccon2.MultiClient = true
					c2, err2 := gipc.StartClient(ccon2)
					if err2 != nil {
						t.Error(err)
					}

					for {
						n, _ := c2.Read()
						if n.Status == "Connected" {
							c2.Close()
							break
						}
					}
					return
				}
			}
		}
	}()

	connect := false
	disconnect := false

	for {
		select {
		case <-hasReconnected:
			return
		default:
			sc.Connections.ReadTimed(2*time.Second, func(_ *gipc.Server, m *gipc.Message, err error) {
				if m.Status == "Connected" && connect == false {
					hasConnected <- true
					connect = true
				}

				if m.Status == "Disconnected" {
					hasDisconnected <- true
					disconnect = true
				}

				if m.Status == "Connected" && connect == true && disconnect == true {
					hasReconnected <- true
				}
			})
		}
	}
}
