package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/jacobsa/go-serial/serial"
	"github.com/sigurn/crc16"
)

type SerialChannel struct {
	queryId   uint32
	Closed    bool
	RW        io.ReadWriteCloser
	writeLock sync.Mutex
	readLock  sync.Mutex
	callbacks map[uint32]chan []byte
}

type SerialFrame struct {
	ChipID uint8
	Data   []byte
}

const (
	STX                  byte = 0x02
	ETX                  byte = 0x03
	ESC                  byte = 0x1B
	PacketHeaderLength        = 5
	PacketChecksumLength      = 2
)

func escape(data []byte) []byte {
	var buf bytes.Buffer
	for _, b := range data {
		switch b {
		case STX:
			fallthrough
		case ESC:
			fallthrough
		case ETX:
			buf.WriteByte(ESC)
			fallthrough
		default:
			buf.WriteByte(b)
		}
	}
	return buf.Bytes()
}

func calcChecksum(data []byte) []byte {
	arr := make([]byte, 2)
	table := crc16.MakeTable(crc16.CRC16_ARC)
	checksum := crc16.Checksum(data, table)
	binary.BigEndian.PutUint16(arr, checksum)
	return arr
}

type packetHeader struct {
	Version uint8
	Type    uint8
	ID      uint8
	Length  uint16
}

func pack(id uint8, requestType uint8, data []byte) []byte {

	// Frame
	header := packetHeader{
		Version: 0,
		Type:    requestType,
		ID:      id,
		Length:  uint16(len(data)),
	}
	var payload bytes.Buffer
	binary.Write(&payload, binary.BigEndian, &header)
	payload.Write(data)
	payload.Write(calcChecksum(payload.Bytes()))
	body := payload.Bytes()
	body = escape(body)

	// Transfer package
	var res bytes.Buffer
	res.WriteByte(STX)
	res.Write(body)
	res.WriteByte(ETX)
	return res.Bytes()
}

func SerialOpen(path string) (*SerialChannel, error) {
	res, err := serial.Open(serial.OpenOptions{
		PortName:              path,
		BaudRate:              115200,
		DataBits:              8,
		StopBits:              2,
		InterCharacterTimeout: 100,
		MinimumReadSize:       0,
		RTSCTSFlowControl:     false,
	})
	if err != nil {
		return nil, err
	} else {
		return &SerialChannel{RW: res, Closed: false, queryId: 0, callbacks: make(map[uint32]chan []byte)}, nil
	}
}

func (channel *SerialChannel) Write(chipId int, data []byte) error {

	// Write lock
	channel.writeLock.Lock()
	defer channel.writeLock.Unlock()

	// Write cycle
	packed := pack(uint8(chipId), 0x0, data)
	n, err := channel.RW.Write(packed)
	if err != nil {
		return err
	}
	if n != len(packed) {
		return errors.New("UART wirte issue")
	}
	return nil
}

func parsePacket(p []byte) (*SerialFrame, error) {
	// check length
	if len(p) < PacketHeaderLength {
		return nil, errors.New("invalid packet")
	}
	// parse header
	hdr := packetHeader{}
	headerBuf := bytes.NewBuffer(p[:PacketHeaderLength])
	binary.Read(headerBuf, binary.BigEndian, &hdr)
	data := p[PacketHeaderLength:]
	if len(data) != int(hdr.Length) {
		return nil, errors.New("invalid packet")
	}
	res := SerialFrame{ChipID: hdr.ID, Data: data}
	return &res, nil
}

func unserialize(p []byte) (*SerialFrame, error) {
	// check crc first
	payload, err := popCRC(p)
	if err != nil {
		return nil, err
	}
	// parse header + check data
	res, err := parsePacket(payload)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func popCRC(p []byte) ([]byte, error) {
	// check length
	if len(p) < PacketChecksumLength {
		return nil, errors.New("invalid packet")
	}
	pcrc := len(p) - PacketChecksumLength
	payload := p[:pcrc]
	checksum := p[pcrc:]
	// checksum
	if !bytes.Equal(checksum, calcChecksum(payload)) {
		return nil, errors.New("invalid packet")
	}
	return payload, nil
}

func (channel *SerialChannel) Read() (*SerialFrame, error) {

	// Read lock
	channel.readLock.Lock()
	defer channel.readLock.Unlock()

	// Read cycle
	var buffer bytes.Buffer
	b := make([]byte, 1)
	for {
		if _, err := channel.RW.Read(b); err != nil {
			return nil, err
		}
		switch b[0] {
		case STX:
			buffer.Reset()
		case ETX:
			frm, err := unserialize(buffer.Bytes())
			if err != nil {
				return nil, err
			}
			return frm, nil
		case ESC:
			if _, err := channel.RW.Read(b); err != nil {
				return nil, err
			}
			fallthrough
		default:
			buffer.WriteByte(b[0])
		}
	}
}

func (channel *SerialChannel) Close() {

	// Write lock
	channel.writeLock.Lock()
	defer channel.writeLock.Unlock()

	// Closing
	channel.Closed = true
	channel.RW.Close()
}
