package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/jacobsa/go-serial/serial"
)

type SerialChannel struct {
	queryId uint32
	Closed  bool
	RW      io.ReadWriteCloser
}

func (channel *SerialChannel) PerformJob(data []byte) {
	queryId := channel.queryId
	channel.queryId++

	// Job Header
	job := []byte{0xa4, 0x61, 0xa1, 0x8c}
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, queryId)

	// Job body
	job = append(job, data...)

	// Sending
	channel.RW.Write(job)

	// Log result
	fmt.Printf("Sent job %d", queryId)
}

func (channel *SerialChannel) Start() {
	go (func() {

		// Read header
		header := make([]byte, 4)
		r, e := channel.RW.Read(header)
		if r != 4 {
			log.Panic("Invalid header length")
		}
		if e != nil {
			log.Panic(e)
		}
		if header[0] != 0x42 || header[1] != 94 || header[2] != 37 || header[3] != 0x9b {
			log.Panic("Invalid header")
		}

		// Read data
		data := make([]byte, 100)
		r, e = channel.RW.Read(data)
		if r != 100 {
			log.Panic("Invalid data length")
		}
		if e != nil {
			log.Panic(e)
		}

		// Parse packet
		jobId := binary.BigEndian.Uint32(data)
		hash := data[4 : 32+4]
		odata := data[32+4:]

		// Print result
		fmt.Printf("Received %d: %x | %x", jobId, hash, odata)
	})()
}

func (channel *SerialChannel) Close() {
	channel.Closed = true
	channel.RW.Close()
}

func SerialOpen(path string) (*SerialChannel, error) {
	res, err := serial.Open(serial.OpenOptions{
		PortName: path,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 2,
		// mode
		InterCharacterTimeout: 100,
		MinimumReadSize:       0,

		RTSCTSFlowControl: false,
	})
	if err != nil {
		return nil, err
	} else {
		return &SerialChannel{RW: res, Closed: false, queryId: 0}, nil
	}
}
