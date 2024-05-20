package gipc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// checks the name passed into the start function to ensure it's ok/will work.
func checkIpcName(ipcName string) error {

	if len(ipcName) == 0 {
		return errors.New("ipcName cannot be an empty string")
	}

	return nil
}

func intToBytes(mLen int) []byte {

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(mLen))

	return b
}

func bytesToInt(b []byte) int {

	var mlen uint32

	binary.Read(bytes.NewReader(b[:]), binary.BigEndian, &mlen) // message length

	return int(mlen)
}

func getLogrusLevel(logLevel string) logrus.Level {
	debugEnv := os.Getenv("GIPC_DEBUG")
	if len(debugEnv) > 0 {
		if debugEnv == "true" {
			return logrus.DebugLevel
		} else {
			return strToLogLevel(debugEnv)
		}
	} else {
		return strToLogLevel(logLevel)
	}
}

func strToLogLevel(str string) logrus.Level {
	switch str {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return DEFAULT_LOG_LEVEL
	}
}

func GetDefaultClientConnectWait() int {
	envVar := os.Getenv("GIPC_WAIT")
	if len(envVar) > 0 {
		valInt, err := strconv.Atoi(envVar)
		if err == nil {
			return valInt
		}
	}
	return DEFAULT_WAIT
}

// Sleep change the sleep time by using GIPC_WAIT env variable (seconds)
func Sleep() {
	wait := GetDefaultClientConnectWait()
	if wait > 5 {
		time.Sleep(time.Duration(wait) * time.Millisecond)
	} else {
		time.Sleep(time.Duration(wait) * time.Second)
	}
}
