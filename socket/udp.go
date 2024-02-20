package socket

import (
	"net"
	"time"
)

const MTU = 1400

type Udp struct {
	timeout time.Duration
	conn    net.Conn
	buffer  [MTU]byte
	wait    time.Duration
	next    time.Time
	addr    string
}

func NewUdp(address string, timeout time.Duration) (*Udp, error) {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &Udp{timeout: timeout, addr: address, conn: conn}, nil
}

func (u *Udp) Connect() error {
	u.Close()
	c, err := net.Dial("udp", u.addr)
	if err != nil {
		return err
	}

	u.conn = c
	return nil
}

func (u *Udp) SetTimeout(duration time.Duration) {
	u.timeout = duration
}

func (u *Udp) RemoteAddr() net.Addr {
	return u.conn.RemoteAddr()
}

func (u *Udp) SetRateLimit(ratePerMinute int) {
	u.wait = (time.Minute / time.Duration(ratePerMinute)) + time.Second
}

func (u *Udp) extendedDeadline() time.Time {
	return time.Now().Add(u.timeout)
}

func (u *Udp) enforceRateLimit() {
	if u.wait == 0 {
		return
	}

	wait := time.Until(u.next)
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (u *Udp) setNextQueryTime() {
	if u.wait != 0 {
		u.next = time.Now().Add(u.wait)
	}
}

func (u *Udp) Send(data []byte) error {
	u.enforceRateLimit()
	defer u.setNextQueryTime()

	if u.timeout > 0 {
		u.conn.SetWriteDeadline(u.extendedDeadline())
	}

	_, err := u.conn.Write(data)
	return err
}

func (u *Udp) Receive() ([]byte, error) {
	defer u.setNextQueryTime()

	if u.timeout > 0 {
		u.conn.SetReadDeadline(u.extendedDeadline())
	}

	n, err := u.conn.Read(u.buffer[0:MTU])
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, n)
	copy(buffer, u.buffer[:n])

	return buffer, nil
}

func (u *Udp) Close() error {
	return u.conn.Close()
}
