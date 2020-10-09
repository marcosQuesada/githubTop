package http

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestNewHTTPServerWorkflow(t *testing.T) {
	s := New(8888, nil, nil, "fakeApp")

	go func() {
		_ = s.Run()
	}()

	c, err := tryConnect(3)
	if err != nil {
		t.Fatalf("Failure on connection to server, err %s", err.Error())
	}

	_ = c.Close()

	s.Terminate()
}

func tryConnect(maxConnRetries int) (net.Conn, error) {
	maxConnRetries--

	c, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		if maxConnRetries == 0 {
			return nil, errors.New("Max Connection Retries done")
		}

		time.Sleep(time.Millisecond * 100)
		_, _ = tryConnect(maxConnRetries)
	}

	return c, nil
}
