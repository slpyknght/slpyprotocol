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
	SEQ_ID int = 1
	SEQ_LENGTH int = 1
	SEQ_INDEX int = 1
	CONTENT_LENGTH int = 2
	
	HEADER_SIZE int = 15
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
	binary.BigEndian.PutUint32(seqId, rand.Uint32())
	var maxLen uint16 = uint16(BUFFER_SIZE - HEADER_SIZE)
	var seqLen = byte(msg.ContentLength / maxLen)
	if msg.ContentLength % maxLen != 0{
		seqLen += 1
	}
	seqLen = max(seqLen, 1)
	fmt.Printf("content: %v, maxLen: %v = %v segments.\n", msg.ContentLength, maxLen, seqLen)
	msg.SequenceId = seqId[0]
	for i := byte(0); i < seqLen; i++{
		r = append(r, createSegment(msg, i, maxLen, seqLen))
	}
	return r
}

func createSegment(msg Message, idx byte, maxLen uint16, seqLen byte)MessageRaw{
	contentStart := uint16(idx) * maxLen
	clen := min(msg.ContentLength - contentStart, maxLen)
	log.Printf("create seg: %v from %v to %v", idx, contentStart, contentStart + clen  )
	segment := []byte{msg.Version, byte(msg.Type)	}
	segment = binary.BigEndian.AppendUint32(segment, msg.Origin)
	segment = binary.BigEndian.AppendUint32(segment, msg.Destination)
	segment = append(segment, msg.SequenceId, seqLen, idx + 1)
	segment = binary.BigEndian.AppendUint16(segment, clen)
	log.Println(segment)
	log.Println(len(segment))
	if clen == 0{
		return segment
	}
	segment = append(segment, msg.Content[contentStart:contentStart+clen]...)
	log.Printf("segment len with content: %d", len(segment))
	return segment
}

func (data MessageRaw)ToMessage()(Message,error){
	if data[0] != 0x01{
		return Message{}, fmt.Errorf("invalid message version: %v", data[0:HEADER_SIZE+1])
	}
	msg := Message{Version: data[0]}
	msg.Type = MessageType(data[1])
	msg.Origin = binary.BigEndian.Uint32(data[2:6])
	msg.Destination = binary.BigEndian.Uint32(data[6:10])
	// msg.Datetime = binary.BigEndian.Uint32(data[10:14])
	msg.SequenceId = data[10]
	msg.SequenceLength = data[11]
	msg.SequenceIndex = data[12]
	// fmt.Printf("bytes:%v, uint16: %v\n", data[13:15], binary.BigEndian.Uint16(data[13:15]))
	msg.ContentLength = binary.BigEndian.Uint16(data[13:15])
	if msg.ContentLength > 0{
		msg.Content = data[15:15+msg.ContentLength]
	}
	return msg, nil
}

func (buffer *MessageBuffer)Add(message Message)(Message, bool, error){
	if message.SequenceLength <= 1{
		return message, true, nil
	}
	var msg Message = message
	*buffer = append(*buffer, msg)
	// log.Printf("recived partial message: %v of %v", msg.SequenceIndex, msg.SequenceLength)
	// log.Printf("got %v of %v", len(buffer), msg.SequenceLength)
	if len(*buffer) == int(msg.SequenceLength){
		for _, x := range *buffer{
			fmt.Println(x)
		}
		m, err := buffer.Combine()
		if err != nil{
			log.Println(err)
			return Message{}, false, err
		}
		msg = m
		return msg, true, nil
	}else{
		return msg, false, nil
	}
}

func (buf *MessageBuffer)Combine()(Message, error){
	if len(*buf) != int((*buf)[0].SequenceLength){
		return Message{}, fmt.Errorf("invalid MessageBuffer length. expected:%d, got:%d", (*buf)[0].SequenceLength, len(*buf))
	}
	m := (*buf)[0]
	m.ContentLength = 0
	m.Content = make([]byte, 0)
	i := byte(0)
	next := byte(1)
	missingIndex := false
	for{
		if (*buf)[i].SequenceIndex == next{
			log.Printf("appending %v of %v", next, m.SequenceLength)
			m.Content = append(m.Content, (*buf)[i].Content...)
			m.ContentLength += (*buf)[i].ContentLength
			missingIndex = false
			next++
			if next > m.SequenceLength{
				break
			}
		}
		i++
		if i == byte(len(*buf)){
			if missingIndex{
				break
			}
			missingIndex = true
			i = 0
		}
	}
	if missingIndex{
		return Message{}, fmt.Errorf("sequence %v is missing index: %v", m.SequenceId, next)
	}
	return m,nil
}

func (m Message)Log(){
	log.Printf("\t\t=== %v ===", m.SequenceId)
	log.Printf("\tversion:\t%v", m.Version)
	log.Printf("\ttype:\t\t%v", m.Type)
	log.Printf("\torigin:\t\t%v", m.Origin)
	log.Printf("\tdest:\t\t%v", m.Destination)
	log.Printf("\tseq:\t\t%v: (%v/%v)", m.SequenceId, m.SequenceIndex, m.SequenceLength)
	log.Printf("\tcontent len:\t%v", m.ContentLength)
}
