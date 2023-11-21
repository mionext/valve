package valve

import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"

	"github.com/mionext/valve/socket"
	"github.com/mionext/valve/types"
)

var ErrBadPacketHeader = errors.New("bad packet header")
var ErrUnknownInfoVersion = errors.New("unknown A2S_INFO version")

// Info gets the server info
func (c *Client) Info() (*types.Server, error) {
	var pb socket.PacketBuilder
	pb.WriteBytes([]byte{0xff, 0xff, 0xff, 0xff, 0x54})
	pb.WriteCString("Source Engine Query")

	if err := c.socket.Send(pb.Bytes()); err != nil {
		return nil, err
	}

	data, err := c.socket.Receive()
	if err != nil {
		return nil, err
	}

	// need challenge
	if data[4] == 0x41 {
		pb.WriteBytes([]byte{data[5], data[6], data[7], data[8]})
		if err := c.socket.Send(pb.Bytes()); err != nil {
			return nil, err
		}

		data, err = c.socket.Receive()
		if err != nil {
			return nil, err
		}
	}

	r := socket.NewPacketReader(data)
	if r.ReadUint32() != 0xFFFFFFFF {
		return nil, ErrBadPacketHeader
	}

	responseType := r.ReadUint8()
	switch responseType {
	case types.GoldSourceInfoResponseType:
		return c.resolveGoldSourceInfo(r)
	case types.SourceInfoResponseType:
		return c.resolveSourceInfo(r)
	}

	return nil, ErrUnknownInfoVersion
}

// resolveGoldSourceInfo resolves the golden server info
func (c *Client) resolveGoldSourceInfo(r *socket.PacketReader) (*types.Server, error) {
	s := &types.Server{Version: types.GoldSourceInfoResponseType, Address: c.socket.RemoteAddr().String()}
	// drop server address
	r.ReadString()
	s.Name = r.ReadString()
	s.Map = r.ReadString()
	s.Folder = r.ReadString()
	s.Game = r.ReadString()
	s.Players = r.ReadUint8()
	s.MaxPlayers = r.ReadUint8()
	s.Protocol = r.ReadUint8()

	serverType := r.ReadUint8()
	switch serverType {
	case uint8('l'):
		s.Type = types.ServerTypeNonDedicated
	case uint8('d'):
		s.Type = types.ServerTypeDedicated
	case uint8('p'):
		s.Type = types.ServerTypeSourceTV
	default:
		s.Type = types.ServerTypeUnknown
	}

	serverOs := r.ReadUint8()
	switch serverOs {
	case uint8('l'):
		s.OS = types.ServerOsLinux
	case uint8('w'):
		s.OS = types.ServerOsWindows
	case uint8('m'):
		s.OS = types.ServerOsMac
	default:
		s.OS = types.ServerOsUnknown
	}

	s.Visibility = r.ReadUint8()

	isMod := r.ReadUint8()
	if isMod == 1 {
		m := &types.ModInfo{}
		m.Url = r.ReadString()
		m.DownloadUrl = r.ReadString()
		r.ReadUint8() // ignore null byte
		m.Version = r.ReadUint32()
		m.Size = r.ReadUint32()
		m.Type = r.ReadUint8()
		m.Dll = r.ReadUint8()
		s.Mod = m
	}

	s.Vac = r.ReadUint8()
	s.Bots = r.ReadUint8()

	return s, nil
}

// resolveSourceInfo resolves the server info
func (c *Client) resolveSourceInfo(r *socket.PacketReader) (*types.Server, error) {
	s := &types.Server{Version: types.SourceInfoResponseType, Address: c.socket.RemoteAddr().String()}
	s.Protocol = r.ReadUint8()
	s.Name = r.ReadString()
	s.Map = r.ReadString()
	s.Folder = r.ReadString()
	s.Game = r.ReadString()

	appId := types.AppId(r.ReadUint16())

	s.Players = r.ReadUint8()
	s.MaxPlayers = r.ReadUint8()
	s.Bots = r.ReadUint8()

	serverType := r.ReadUint8()
	switch serverType {
	case uint8('l'):
		s.Type = types.ServerTypeDedicated
	case uint8('d'):
		s.Type = types.ServerTypeNonDedicated
	case uint8('p'):
		s.Type = types.ServerTypeSourceTV
	default:
		s.Type = types.ServerTypeUnknown
	}

	serverOs := r.ReadUint8()
	switch serverOs {
	case uint8('l'):
		s.OS = types.ServerOsLinux
	case uint8('w'):
		s.OS = types.ServerOsWindows
	case uint8('m'):
		s.OS = types.ServerOsMac
	default:
		s.OS = types.ServerOsUnknown
	}

	s.Visibility = r.ReadUint8()
	s.Vac = r.ReadUint8()

	if appId == types.AppTheShip {
		theShip := &types.TheShip{}
		theShip.Mode = r.ReadUint8()
		theShip.Witnesses = r.ReadUint8()
		theShip.Duration = r.ReadUint8()
		s.TheShip = theShip
	}

	s.Extended = &types.Extended{AppId: appId, GameVersion: r.ReadString()}

	if !r.More() {
		return s, nil
	}

	edf := r.ReadUint8()
	if edf&0x80 != 0 {
		s.Extended.Port = r.ReadUint16()
	}

	if edf&0x10 != 0 {
		s.Extended.SteamId = r.ReadUint64()
	}

	if edf&0x40 != 0 {
		s.SourceTV = &types.SourceTV{Port: r.ReadUint16(), Name: r.ReadString()}
	}

	if edf&0x20 != 0 {
		s.Extended.GameModeDescription = r.ReadString()
	}

	if edf&0x01 != 0 && r.CanRead(8) {
		gameId := r.ReadUint64()
		s.Extended.GameId = gameId
		// bits 0-23: true app id (original could be truncated)
		// bits 24-31: type
		// bits 32-63: mod id
		s.Extended.AppId = types.AppId(gameId & uint64(0xffffffff))
	}

	return s, nil
}

// Players  gets the server players
func (c *Client) Players() (*types.PlayerList, error) {
	req := []byte{0xff, 0xff, 0xff, 0xff, 0x55, 0xff, 0xff, 0xff, 0xff}
	if err := c.socket.Send(req); err != nil {
		return nil, err
	}

	data, err := c.socket.Receive()
	if err != nil {
		return nil, err
	}

	// has challenge
	if data[4] == 0x41 {
		reply := []byte{
			0xff, 0xff, 0xff, 0xff, 0x55,
			data[5], data[6], data[7], data[8],
		}

		if err := c.socket.Send(reply); err != nil {
			return nil, err
		}

		data, err = c.socket.Receive()
		if err != nil {
			return nil, err
		}
	}

	r := socket.NewPacketReader(data)
	if !r.CanRead(4) {
		return nil, errors.New("bad player reply")
	}

	r.ReadInt32() // drop header
	if r.ReadUint8() != 0x44 {
		return nil, errors.New("bad player header")
	}

	count := int(r.ReadUint8())
	players := &types.PlayerList{Count: count}
	for i := 0; i < count; i++ {
		p := &types.Player{}
		p.Index = int(r.ReadUint8())
		p.Name = r.ReadString()
		p.Score = int(r.ReadUint32())
		p.Duration = int(r.ReadFloat32())
		players.Players = append(players.Players, p)
	}

	return players, nil
}

func (c *Client) Rules() (map[string]string, error) {
	s, err := c.Info()
	if err != nil {
		return nil, err
	}

	c.Reconnect()
	req := []byte{0xff, 0xff, 0xff, 0xff, 0x56, 0x00, 0x00, 0x00, 0x00}
	if err := c.socket.Send(req); err != nil {
		return nil, err
	}

	data, err := c.socket.Receive()
	if err != nil {
		return nil, err
	}

	// has challenge
	if data[4] == 0x41 {
		reply := []byte{
			0xff, 0xff, 0xff, 0xff, 0x56,
			data[5], data[6], data[7], data[8],
		}

		if err := c.socket.Send(reply); err != nil {
			return nil, err
		}

		data, err = c.socket.Receive()
		if err != nil {
			return nil, err
		}
	}

	compressed := false
	switch int32(binary.LittleEndian.Uint32(data[:4])) {
	case -1: // continue
	case -2:
		data, compressed, err = c.waitForMultiPacketReply(data, s)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("bad packet header")
	}

	r := socket.NewPacketReader(data)
	if compressed {
		decompressedSize := r.ReadUint32()
		checksum := r.ReadUint32()
		if decompressedSize > uint32(1024*1024) {
			return nil, errors.New("decompressed size too large")
		}

		decompressed := make([]byte, decompressedSize)
		bz2Reader := bzip2.NewReader(bytes.NewReader(data[r.Pos():]))
		n, err := bz2Reader.Read(decompressed)
		if err != nil {
			return nil, err
		}

		if n != int(decompressedSize) {
			return nil, errors.New("decompressed size mismatch")
		}

		if crc32.ChecksumIEEE(decompressed) != checksum {
			return nil, errors.New("checksum mismatch")
		}

		data = decompressed
		r = socket.NewPacketReader(data)
	}

	if r.ReadInt32() != -1 {
		return nil, errors.New("bad packet header")
	}

	if r.ReadUint8() != 0x45 {
		return nil, errors.New("bad rules reply")
	}

	rules := map[string]string{}
	count := int(r.ReadUint16())
	for i := 0; i < count; i++ {
		key, err := r.TryReadString()
		if err != nil {
			break
		}

		value, err := r.TryReadString()
		if err != nil {
			break
		}

		rules[key] = value
	}

	return rules, nil
}

// MultiPacketHeader represents a multi packet header
func (c *Client) waitForMultiPacketReply(data []byte, s *types.Server) ([]byte, bool, error) {
	header, err := c.decodeMultiPacketHeader(data, s)
	if err != nil {
		return nil, false, err
	}

	packets := make([]*MultiPacketHeader, header.TotalPackets)
	received := 0
	fullSize := 0

	for {
		if int(header.PacketNumber) >= len(packets) {
			return nil, false, errors.New("bad packet number")
		}
		if packets[header.PacketNumber] != nil {
			return nil, false, errors.New("duplicate packet")
		}

		packets[header.PacketNumber] = header
		fullSize += len(header.Payload)
		received++

		if received == len(packets) {
			break
		}

		data, err := c.socket.Receive()
		if err != nil {
			return nil, false, err
		}

		header, err = c.decodeMultiPacketHeader(data, s)
		if err != nil {
			return nil, false, err
		}
	}

	payloads := make([]byte, fullSize)
	cursor := 0
	for _, header := range packets {
		copy(payloads[cursor:cursor+len(header.Payload)], header.Payload)
		cursor += len(header.Payload)
	}

	return payloads, packets[0].Compressed, nil
}

func (c *Client) decodeMultiPacketHeader(data []byte, s *types.Server) (*MultiPacketHeader, error) {
	r := socket.NewPacketReader(data)
	if r.ReadInt32() != -2 {
		return nil, errors.New("not a multi packet header")
	}

	header := &MultiPacketHeader{}
	header.Id = r.ReadUint32()
	switch s.Engine() {
	case types.GoldSource:
		pkt := r.ReadUint8()
		header.PacketNumber = (pkt >> 4) & 0x0F
		header.TotalPackets = pkt & 0x0F
	case types.Source:
		header.Compressed = (header.Id & uint32(0x80000000)) != 0
		header.TotalPackets = r.ReadUint8()
		header.PacketNumber = r.ReadUint8()
		if r.CanRead(2) {
			header.PacketSize = r.ReadUint16()
		}
	default:
		return nil, errors.New("unknown game engine")
	}

	header.Size = r.Pos()
	header.Payload = data[header.Size:]

	return header, nil
}

// Ping gets the server ping in milliseconds
func (c *Client) Ping() (uint, error) {
	start := time.Now()
	data := []byte{0xff, 0xff, 0xff, 0xff, 0x69}
	if err := c.socket.Send(data); err != nil {
		return 0, err
	}

	data, err := c.socket.Receive()
	if err != nil {
		return 0, err
	}

	if data[4] != 0x6a {
		return 0, errors.New("bad ping reply")
	}

	return uint(time.Since(start).Milliseconds()), nil
}
