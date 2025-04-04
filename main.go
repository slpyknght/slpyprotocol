package slpyprotocol

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

const(
	// length in byte
	VERSION int = 1
	TYPE int = 1
	ORIGIN int = 4
	DESTINATION int = 4
	SEQ_LENGTH int = 1
	SEQ_INDEX int = 1
	CONTENT_LENGTH int = 4
	
	HEADER_SIZE int = 16
	BUFFER_SIZE int = 128
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
type MessageBuffer []Message;

type Message struct{
	Version byte
	Type MessageType
	Origin uint32
	Destination uint32
	// Datetime uint32
	SequenceId byte
	SequenceLength byte
	SequenceIndex byte
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
		ContentLength: 0,
		SequenceIndex: 1,
		SequenceLength: 1,
		SequenceId: 1,
	}
}

func (msg Message)ToPackage()[]MessageRaw{
	r := make([]MessageRaw,0)
	// loop and divide content if > 1024-header bytes to fit into buffer
	// if msg.ContentLength > uint16(maxLen){
	// }
	// contentIdx := 0
	seqId := make([]byte,4)
	binary.LittleEndian.PutUint32(seqId, rand.Uint32())
	var maxLen uint16 = uint16(BUFFER_SIZE - HEADER_SIZE) + 1
	remainingContent := msg.ContentLength
	var cidx uint16 = 0
	var seqLen = 1 + byte(msg.ContentLength / maxLen)
	for i := byte(0); i < seqLen; i++{
		var clen uint16
		clen = min(maxLen, remainingContent)
		a := []byte{msg.Version, byte(msg.Type)	}
		a = binary.LittleEndian.AppendUint32(a, msg.Origin)
		a = binary.LittleEndian.AppendUint32(a, msg.Destination)
		// a = binary.LittleEndian.AppendUint32(a, 1) // time or packacke number
		a = append(a, seqId[0], seqLen, (i+1))
		// fmt.Printf("content len %v / %v",clen, binary.LittleEndian.(clen))
		a = binary.LittleEndian.AppendUint16(a, clen)
		a = append(a, msg.Content[cidx:cidx+clen]...)
		r = append(r, a)
		cidx += clen
		remainingContent -= clen
		fmt.Printf("created package[%v/%v]: %v\n", i+1, seqLen, len(a))
		fmt.Printf("package header: %v\n", a[:HEADER_SIZE])
	}
	return r
}

func (data MessageRaw)ToMessage()(Message,error){
	if data[0] != 0x01{
		return Message{}, fmt.Errorf("invalid message version: %v", data[0])
	}
	msg := Message{Version: data[0]}
	msg.Type = MessageType(data[1])
	msg.Origin = binary.LittleEndian.Uint32(data[2:6])
	msg.Destination = binary.LittleEndian.Uint32(data[6:10])
	// msg.Datetime = binary.LittleEndian.Uint32(data[10:14])
	msg.SequenceId = data[10]
	msg.SequenceLength = data[11]
	msg.SequenceIndex = data[12]
	// fmt.Printf("bytes:%v, uint16: %v\n", data[13:15], binary.LittleEndian.Uint16(data[13:15]))
	msg.ContentLength = binary.LittleEndian.Uint16(data[13:15])
	if msg.ContentLength > 0{
		msg.Content = data[15:15+msg.ContentLength]
	}
	return msg, nil
}

func (buf MessageBuffer)Combine()(Message, error){
	if len(buf) != int(buf[0].SequenceLength){
		return Message{}, fmt.Errorf("invalid MessageBuffer length. expected:%d, got:%d", buf[0].SequenceLength, len(buf))
	}
	m := Message{
		Version: buf[0].Version,
		Type: buf[0].Type,
		Origin: buf[0].Origin,
		Destination: buf[0].Destination,
		SequenceId: buf[0].SequenceId,
		SequenceLength: buf[0].SequenceLength,
		ContentLength: 0,
		Content: make([]byte, 0),
	}
	i := byte(0)
	next := byte(1)
	missingIndex := false
	for{
		if buf[i].SequenceIndex == next{
			log.Printf("appending %v of %v", next, m.SequenceLength)
			m.Content = append(m.Content, buf[i].Content...)
			m.ContentLength += buf[i].ContentLength
			missingIndex = false
			next++
			if next >= m.SequenceLength{
				break
			}
		}
		i++
		if i == byte(len(buf)){
			if missingIndex{
				break
			}
			missingIndex = true
			i = 0
		}
	}
	if missingIndex{
		return Message{}, fmt.Errorf("sequence is missing an index: %v", next)
	}
	return m,nil
}

