package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"gitlab.neoclub.cn/NeoGo/gnet"
)

// CustomLengthFieldProtocol : custom protocol
// custom protocol header contains Version, ActionType and DataLength fields
// its payload is Data field
type CustomLengthFieldProtocol struct {
	Version    uint16
	ActionType uint16
	DataLength uint32
	Data       []byte
}

// Encode ...
func (cc *CustomLengthFieldProtocol) Encode(dataFrame interface{}) ([]byte, error) {
	result := make([]byte, 0)

	item := dataFrame.(*CustomLengthFieldProtocol)

	buffer := bytes.NewBuffer(result)

	if err := binary.Write(buffer, binary.BigEndian, item.Version); err != nil {
		s := fmt.Sprintf("Pack version error , %v", err)
		return nil, errors.New(s)
	}

	if err := binary.Write(buffer, binary.BigEndian, item.ActionType); err != nil {
		s := fmt.Sprintf("Pack type error , %v", err)
		return nil, errors.New(s)
	}
	dataLen := uint32(len(item.Data))
	if err := binary.Write(buffer, binary.BigEndian, dataLen); err != nil {
		s := fmt.Sprintf("Pack datalength error , %v", err)
		return nil, errors.New(s)
	}
	if dataLen > 0 {
		if err := binary.Write(buffer, binary.BigEndian, item.Data); err != nil {
			s := fmt.Sprintf("Pack data error , %v", err)
			return nil, errors.New(s)
		}
	}

	return buffer.Bytes(), nil
}

// Decode ...
func (cc *CustomLengthFieldProtocol) Decode(c gnet.Conn, dataFrame interface{}) (interface{}, error) {
	obj := dataFrame.(*CustomLengthFieldProtocol)

	// parse header
	headerLen := DefaultHeadLength // uint16+uint16+uint32
	if size, header := c.ReadN(headerLen); size == headerLen {
		byteBuffer := bytes.NewBuffer(header)
		_ = binary.Read(byteBuffer, binary.BigEndian, &obj.Version)
		_ = binary.Read(byteBuffer, binary.BigEndian, &obj.ActionType)
		_ = binary.Read(byteBuffer, binary.BigEndian, &obj.DataLength)
		// to check the protocol version and actionType,
		// reset buffer if the version or actionType is not correct
		if obj.Version != DefaultProtocolVersion || isCorrectAction(obj.ActionType) == false {
			c.ResetBuffer()
			log.Println("not normal protocol:", obj.Version, DefaultProtocolVersion, obj.ActionType, obj.DataLength)
			return nil, errors.New("not normal protocol")
		}
		// parse payload
		dataLen := int(obj.DataLength) // max int32 can contain 210MB payload
		protocolLen := headerLen + dataLen
		if dataSize, data := c.ReadN(protocolLen); dataSize == protocolLen {
			c.ShiftN(protocolLen)
			// log.Println("parse success:", data, dataSize)

			// return the payload of the data
			obj.Data = data[headerLen:]
			return obj, nil
		}
		// log.Println("not enough payload data:", dataLen, protocolLen, dataSize)
		return obj, errors.New("not enough payload data")

	}
	// log.Println("not enough header data:", size)
	return obj, errors.New("not enough header data")
}


// its payload is Data field
type NimProtocol struct {
	HeaderLength    uint16
	CmdId 			uint16
	BodyLength 		uint16
	Body       		[]byte
}

// Decode ...
func (cc *NimProtocol) Decode(c gnet.Conn, dataFrame interface{}) (interface{}, error) {
	obj := dataFrame.(*NimProtocol)

	// parse header
	if size, headerLenBytes := c.ReadN(DefaultHeadLengthBytes); size == DefaultHeadLengthBytes {
		obj.HeaderLength = binary.BigEndian.Uint16(headerLenBytes)

		headerLen := int(obj.HeaderLength)
		if size, header := c.ReadN(headerLen); size == headerLen {
			byteBuffer := bytes.NewBuffer(header[DefaultHeadLengthBytes:])
			_ = binary.Read(byteBuffer, binary.BigEndian, &obj.CmdId)
			_ = binary.Read(byteBuffer, binary.BigEndian, &obj.BodyLength)

			if isCorrectAction(obj.CmdId) == false {
				c.ResetBuffer()
				log.Println("not normal protocol:", obj.CmdId)
				return nil, errors.New("not normal protocol")
			}

			// parse payload
			dataLen := int(obj.BodyLength) // max uint16 can contain 210MB payload
			protocolLen := headerLen + dataLen
			if dataSize, data := c.ReadN(protocolLen); dataSize == protocolLen {
				c.ShiftN(protocolLen)
				// log.Println("parse success:", data, dataSize)

				// return the payload of the data
				obj.Body = data[headerLen:]
				return obj, nil
			}
			// log.Println("not enough payload data:", dataLen, protocolLen, dataSize)
			return obj, errors.New("not enough payload data")
		}
	}
	// log.Println("not enough header data:", size)
	return obj, errors.New("not enough header data")
}


// Encode ...
func (cc *NimProtocol) Encode(dataFrame interface{}) ([]byte, error) {
	result := make([]byte, 0)

	item := dataFrame.(*NimProtocol)

	buffer := bytes.NewBuffer(result)

	if err := binary.Write(buffer, binary.BigEndian, uint16(DefHeadLength)); err != nil {
		s := fmt.Sprintf("Pack version error , %v", err)
		return nil, errors.New(s)
	}

	if err := binary.Write(buffer, binary.BigEndian, item.CmdId); err != nil {
		s := fmt.Sprintf("Pack type error , %v", err)
		return nil, errors.New(s)
	}
	dataLen := uint16(len(item.Body))
	if err := binary.Write(buffer, binary.BigEndian, dataLen); err != nil {
		s := fmt.Sprintf("Pack datalength error , %v", err)
		return nil, errors.New(s)
	}
	if dataLen > 0 {
		if err := binary.Write(buffer, binary.BigEndian, item.Body); err != nil {
			s := fmt.Sprintf("Pack data error , %v", err)
			return nil, errors.New(s)
		}
	}

	fmt.Println(buffer.Bytes())
	return buffer.Bytes(), nil
}



// default custom protocol const
const (
	DefaultHeadLengthBytes = 2
	DefHeadLength = 6

	DefaultHeadLength = 8

	DefaultProtocolVersion = 0x8001 // test protocol version

	ActionPing = 0x0001 // ping
	ActionPong = 0x0002 // pong
	ActionData = 0x00F0 // business
)

func isCorrectAction(actionType uint16) bool {
	switch actionType {
	case ActionPing, ActionPong, ActionData:
		return true
	default:
		return false
	}
}

func isCorrectCmdId(cmdId uint16) bool {
	switch cmdId {
	case ActionPing, ActionPong, ActionData:
		return true
	default:
		return false
	}
}


