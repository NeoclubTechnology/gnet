package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/toury12/gnet/test/protocol"
)

// Example command: go run client.go
func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go func() {
		for i := 0; i< 10; i++ {

			response, err := ClientDecode(conn)
			if err != nil {
				log.Printf("ClientDecode error, %v\n", err)
			}

			log.Printf("receive , %v, data:%s\n", response, string(response.Body))

		}
	}()

	data := []byte("hello")
	pbdata, err := ClientEncode(0x00F0, data)
	if err != nil {
		panic(err)
	}
	conn.Write(pbdata)

	data = []byte("world")
	pbdata, err = ClientEncode(0x00F0, data)
	if err != nil {
		panic(err)
	}
	conn.Write(pbdata)

	select {}
}

// ClientEncode :
func ClientEncode(cmdId uint16, data []byte) ([]byte, error) {
	result := make([]byte, 0)

	buffer := bytes.NewBuffer(result)

	if err := binary.Write(buffer, binary.BigEndian, uint16(6)); err != nil {
		s := fmt.Sprintf("Pack version error , %v", err)
		return nil, errors.New(s)
	}

	if err := binary.Write(buffer, binary.BigEndian, cmdId); err != nil {
		s := fmt.Sprintf("Pack type error , %v", err)
		return nil, errors.New(s)
	}
	dataLen := uint16(len(data))
	if err := binary.Write(buffer, binary.BigEndian, dataLen); err != nil {
		s := fmt.Sprintf("Pack datalength error , %v", err)
		return nil, errors.New(s)
	}
	if dataLen > 0 {
		if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
			s := fmt.Sprintf("Pack data error , %v", err)
			return nil, errors.New(s)
		}
	}

	fmt.Println(buffer.Bytes())
	return buffer.Bytes(), nil
}

// ClientDecode :
func ClientDecode(rawConn net.Conn) (*protocol.NimProtocol, error) {
	newPackage := protocol.NimProtocol{}

	headData := make([]byte, protocol.DefHeadLength)
	n, err := io.ReadFull(rawConn, headData)
	if n != protocol.DefHeadLength {
		return nil, err
	}

	// parse protocol header
	bytesBuffer := bytes.NewBuffer(headData)
	binary.Read(bytesBuffer, binary.BigEndian, &newPackage.HeaderLength)
	binary.Read(bytesBuffer, binary.BigEndian, &newPackage.CmdId)
	binary.Read(bytesBuffer, binary.BigEndian, &newPackage.BodyLength)

	if newPackage.BodyLength < 1 {
		return &newPackage, nil
	}

	data := make([]byte, newPackage.BodyLength)
	dataNum, err2 := io.ReadFull(rawConn, data)
	if uint16(dataNum) != newPackage.BodyLength {
		s := fmt.Sprintf("read data error, %v", err2)
		return nil, errors.New(s)
	}

	newPackage.Body = data

	return &newPackage, nil
}
