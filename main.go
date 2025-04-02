package slpyprotocol

import (
	"encoding/binary"
	"fmt"
)

const(
	VERSION_LEN int = 1
	DATETIME_LEN int = 4
	ID_LEN int = 1
	TYPE_LEN int = 1
	CONTENT_LEN int = 1
)

type MessageType int

const (
	MT_None MessageType = iota
	MT_Handshake
	MT_Ping
	MT_Connect
	MT_Disconnect
	MT_Msg
	MT_File
)

type Message struct{
	Version byte
	Type MessageType
	Sender uint32
	Content []byte
}

func CreateMessage(data []byte)(Message,error){
	if data[0] != 0x01{
		return Message{}, fmt.Errorf("invalid message version: %v", data[0])
	}
	msg := Message{Version: data[0]}
	msg.Type = MessageType(data[1])
	msg.Sender = binary.LittleEndian.Uint32(data[2:6])
	msg.Content = data[6:]
	return msg, nil
}


