package valve

import (
	"fmt"
	"time"

	"github.com/oxzz/valve/socket"
)

// Client is a client for the Valve source query protocol
type Client struct {
	socket    *socket.Udp
	timeout   time.Duration
	connected bool
}

// NewClient creates a new client
func NewClient(addr string, timeout time.Duration) (*Client, error) {
	socket, err := socket.NewUdp(addr, timeout)
	if err != nil {
		return nil, err
	}

	return &Client{socket: socket, timeout: timeout, connected: true}, nil
}

// Close closes the underlying socket
func (c *Client) Close() error {
	if c.connected {
		c.connected = false
		return c.socket.Close()
	}

	return nil
}

func (c *Client) Reconnect() error {
	c.Close()
	return c.Connect()
}

func (c *Client) Connect() error {
	if c.connected {
		return nil
	}

	err := c.socket.Connect()
	if err != nil {
		c.connected = false
		return err
	}

	c.connected = true
	return nil
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

type MultiPacketHeader struct {
	// Size of the packet header itself.
	Size int

	// Packet sequence id.
	Id uint32

	// Packet number out of Total Packets.
	PacketNumber uint8

	// Total number of packets to receive.
	TotalPackets uint8

	// Packet size (0 if not present).
	PacketSize uint16

	// Compression information.
	Compressed bool

	Payload []byte
}
