package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/sigurn/crc16"
)

type SerialChannel struct {
	Tag         string
	queryIdLock sync.Mutex
	queryId     uint32
	Closed      bool
	RW          io.ReadWriteCloser
	requestLock sync.Mutex
	writeLock   sync.Mutex
	readLock    sync.Mutex
	callbacks   map[uint32]chan []byte
}

type SerialFrame struct {
	ChipID uint8
	Data   []byte
}

func SerialOpen(path string, speed uint) (*SerialChannel, error) {
	res, err := serial.Open(serial.OpenOptions{
		PortName:              path,
		BaudRate:              speed,
		DataBits:              8,
		StopBits:              1,
		InterCharacterTimeout: 100,
		MinimumReadSize:       0,
		RTSCTSFlowControl:     false,
	})
	if err != nil {
		return nil, err
	} else {
		return &SerialChannel{RW: res, Closed: false, queryId: 0, callbacks: make(map[uint32]chan []byte), Tag: path}, nil
	}
}

func (channel *SerialChannel) Write(chipId int, reqType uint8, data []byte) error {
	channel.writeLock.Lock()
	defer channel.writeLock.Unlock()
	return channel.doWrite(chipId, reqType, data)
}

func (channel *SerialChannel) Read() (*SerialFrame, error) {
	timer := time.NewTimer(1000 * time.Millisecond)
	doneError := make(chan error, 1)
	doneFrame := make(chan *SerialFrame, 1)
	go func() {
		channel.readLock.Lock()
		defer channel.readLock.Unlock()
		r, err := channel.doRead()
		if err != nil {
			doneError <- err
		} else {
			doneFrame <- r
		}
	}()
	select {
	case err := <-doneError:
		return nil, err
	case p := <-doneFrame:
		return p, nil
	case <-timer.C:
		return nil, errors.New("Request timeout")
	}
}

func (channel *SerialChannel) Request(chipId int, reqType uint8, data []byte) (*SerialFrame, error) {
	channel.requestLock.Lock()
	defer channel.requestLock.Unlock()

	// log.Printf("Request")
	err := channel.Write(chipId, reqType, data)
	if err != nil {
		return nil, err
	}

	// log.Printf("Request read")
	res, err := channel.Read()
	// log.Printf("Request end")

	return res, err
}

func (channel *SerialChannel) Close() {

	// Write lock
	channel.writeLock.Lock()
	defer channel.writeLock.Unlock()

	// Closing
	channel.Closed = true
	channel.RW.Close()
}

//////////////////////////////////////////////////////////////////////////////////////////
//  PLL
//////////////////////////////////////////////////////////////////////////////////////////

const (
	PllWrite = 0x0A
	PllRead  = 0x0B
	PllLock  = 0x0C
)

func (channel *SerialChannel) PllGet(chipId int, addr uint8) (uint16, error) {
	resp, err := channel.Request(chipId, 0xA2, []byte{PllRead, addr})
	if err != nil {
		return 0, err
	}
	value := binary.BigEndian.Uint16(resp.Data)
	return value, nil
}

func (channel *SerialChannel) PllSet(chipId int, addr uint8, value uint16) error {
	req := []byte{PllWrite, addr, 0, 0}
	binary.BigEndian.PutUint16(req[2:], value)
	_, err := channel.Request(chipId, 0xA2, req)
	if err != nil {
		return err
	}
	return nil
}

func (channel *SerialChannel) PllSetMask(chipId int, cv PllConstValue) error {
	oldValue, err := channel.PllGet(chipId, cv.Const.Addr)
	if err != nil {
		return err
	}
	newValue := (oldValue & cv.Const.Mask) | (cv.Value & ^cv.Const.Mask)
	if err = channel.PllSet(chipId, cv.Const.Addr, newValue); err != nil {
		return err
	}
	return nil
}

func (channel *SerialChannel) PllApply(chipId int, cvs []PllConstValue, prop *XilinxProperty) error {
	power, err := channel.PllGet(chipId, prop.PLLPowerAddr)
	if err != nil {
		return err
	}
	if err = channel.PllSet(chipId, prop.PLLPowerAddr, 0xFFFF); err != nil {
		return err
	}
	for _, cv := range cvs {
		if err = channel.PllSetMask(chipId, cv); err != nil {
			return err
		}
	}
	if err = channel.PllSet(chipId, prop.PLLPowerAddr, power); err != nil {
		return err
	}
	return nil
}

func (channel *SerialChannel) SetFrequency(chipId int, frequency int) error {
	setup, found := Xilinx7Series.PLLFreq[frequency]
	if !found {
		return nil
	}
	if err := channel.PllApply(chipId, setup, &Xilinx7Series); err != nil {
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////
//  Implementation
//////////////////////////////////////////////////////////////////////////////////////////

func (channel *SerialChannel) doWrite(chipId int, reqType uint8, data []byte) error {
	packed := pack(uint8(chipId), reqType, data)
	// log.Printf("[%v] Write: %x: %d|%d|%x", channel.Tag, packed, chipId, reqType, data)
	n, err := channel.RW.Write(packed)
	if err != nil {
		return err
	}
	if n != len(packed) {
		return errors.New("UART wirte issue")
	}
	return nil
}

func (channel *SerialChannel) doRead() (*SerialFrame, error) {
	var buffer bytes.Buffer
	var buffer2 bytes.Buffer
	b := make([]byte, 1)
	for {
		n, err := channel.RW.Read(b)
		switch err {
		case io.EOF:
			continue
		case nil:
		default:
			return nil, err
		}
		if n != 1 {
			continue
		}
		// log.Printf("Received: %02x", b[0])
		buffer2.WriteByte(b[0])

		switch b[0] {
		case STX:
			buffer.Reset()
		case ETX:
			// log.Printf("Frame (0): %02x", buffer.Bytes())
			// log.Printf("Frame (r): %02x", buffer2.Bytes())
			frm, err := unserialize(buffer.Bytes())
			if err != nil {
				return nil, err
			}
			// log.Printf("Frame: %02x", frm.Data)
			return frm, nil
		case ESC:
			for {
				n, err := channel.RW.Read(b)
				switch err {
				case io.EOF:
					continue
				case nil:
				default:
					return nil, err
				}
				if n != 1 {
					continue
				}
				buffer2.WriteByte(b[0])
				break
			}
			// log.Printf("Received: %02x", b[0])
			fallthrough
		default:
			buffer.WriteByte(b[0])
		}
	}
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

func parsePacket(p []byte) (*SerialFrame, error) {
	// check length
	if len(p) < PacketHeaderLength {
		return nil, fmt.Errorf("invalid parsing: %x", p)
	}
	// parse header
	hdr := packetHeader{}
	headerBuf := bytes.NewBuffer(p[:PacketHeaderLength])
	binary.Read(headerBuf, binary.BigEndian, &hdr)
	data := p[PacketHeaderLength:]
	if len(data) != int(hdr.Length) {
		return nil, fmt.Errorf("invalid parsing: %x", p)
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
		return nil, fmt.Errorf("invalid packet: %x", p)
	}
	pcrc := len(p) - PacketChecksumLength
	payload := p[:pcrc]
	checksum := p[pcrc:]
	// checksum
	if !bytes.Equal(checksum, calcChecksum(payload)) {
		return nil, fmt.Errorf("checksum failed expected %x, got %x. data: %x", calcChecksum(payload), checksum, p)
	}
	return payload, nil
}
