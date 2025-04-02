package slpyprotocol

import (
	"encoding/binary"
	"fmt"
	"time"
)

const(
	// length in byte
	VERSION int = 1
	TYPE int = 1
	ORIGIN int = 4
	DESTINATION int = 4
	DATETIME int = 4
	CONTENT int = 4
)

type MessageType int

const (
	MT_None MessageType = iota
	MT_Handshake
	MT_Ping
	MT_Connect
	MT_Disconnect
	MT_Msg
	MT_Announcement
	MT_File
)

type MessageRaw []byte;

type Message struct{
	Version byte
	Type MessageType
	Origin uint32
	Destination uint32
	Datetime uint32
	ContentLength uint16
	Content []byte
}

type ConnectionArgs struct{
	Name string `json:"name"`
	Datetime time.Time `json:"datetime"`
}

func NewMessage(dest uint32, p MessageType)Message{
	return Message{
		Version: 1,
		Type: p,
		Origin: 0,
		Destination: dest,
		Datetime: 1,
		ContentLength: 0,
	}
}

func (data MessageRaw)ToMessage()(Message,error){
	if data[0] != 0x01{
		return Message{}, fmt.Errorf("invalid message version: %v", data[0])
	}
	msg := Message{Version: data[0]}
	msg.Type = MessageType(data[1])
	msg.Origin = binary.LittleEndian.Uint32(data[2:6])
	msg.Destination = binary.LittleEndian.Uint32(data[6:10])
	msg.Datetime = binary.LittleEndian.Uint32(data[10:14])
	msg.ContentLength = binary.LittleEndian.Uint16(data[14:16])
	msg.Content = data[16:16+msg.ContentLength]
	return msg, nil
}
func (msg Message)ToPackage()[]MessageRaw{
	r := make([]MessageRaw,0)
	// loop and divide content if > 1024-header bytes to fit into buffer
	a := []byte{msg.Version, byte(msg.Type)	}
	a = binary.LittleEndian.AppendUint32(a, msg.Origin)
	a = binary.LittleEndian.AppendUint32(a, msg.Destination)
	a = binary.LittleEndian.AppendUint32(a, 1) // time or packacke number
	a = binary.LittleEndian.AppendUint16(a, msg.ContentLength)
	a = append(a, msg.Content...)
	r = append(r, a)
	return r
}

