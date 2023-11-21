package valve

import (
	"errors"

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
