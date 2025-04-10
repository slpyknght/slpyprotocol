package slpyprotocol

import "time"

type Message struct{
	baseMessage MessageBase
	Author User
	Channel uint32
	Content string
}

type User struct{
	Id uint32 `json:"id"`
	Name string `json:"name"`
}

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
	MT_Resource
)

type ResourceType byte
const(
	RT_User ResourceType = iota
)

type MessageRaw []byte;
type MessageBuffer []MessageBase;

type MessageBase struct{
	Version byte
	Type MessageType
	Origin uint32
	Destination uint32
	ServerTime uint32
	// Datetime uint32
	SequenceId byte
	SequenceLength byte
	SequenceIndex byte
	ContentLength uint16
	Content []byte
}

type ConnectionArgs struct{
	Name string `json:"name,omitempty"`
	ServerTime time.Time `json:"datetime"`
}

type ResourceArgs struct{
	Type ResourceType `json:"type"`
	Id uint32 `json:"id,omitempty"`
}


