package valve

import (
	"fmt"
	"time"

	"github.com/mionext/valve/socket"
)

// Client is a client for the Valve source query protocol
type Client struct {
	socket  *socket.Udp
	timeout time.Duration
}

// NewClient creates a new client
func NewClient(addr string, timeout time.Duration) (*Client, error) {
	socket, err := socket.NewUdp(addr, timeout)
	if err != nil {
		return nil, err
	}

	return &Client{socket: socket, timeout: timeout}, nil
}

// Close closes the underlying socket
func (c *Client) Close() error {
	return c.socket.Close()
}

func Try(fn func() error) error {
	var outErr error
	(func() {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				outErr = err
			}
		}()

		outErr = fn()
	})()

	return outErr
}
