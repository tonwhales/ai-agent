package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"

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
	n, err := channel.RW.Write(job)
	if err != nil {
		log.Panicln(err)
	}
	if n != len(job) {
		log.Panicln("Invalid write")
	}

	// Log result
	fmt.Printf("Sent job %d\n", queryId)
}

func (channel *SerialChannel) Start() {
	go (func() {

		for {

			// Read header
			header := make([]byte, 4)
			r, e := channel.RW.Read(header)
			if e != nil {
				log.Panic(e)
			}
			if r == 0 {
				time.Sleep(1 * time.Second)
				continue
			}
			if r != 4 {
				log.Panicf("Invalid header length: %d", r)
			}
			if header[0] != 0x42 || header[1] != 0x94 || header[2] != 0x37 || header[3] != 0x9b {
				log.Panicf("Invalid header %x", header)
			}

			// Read data
			data := make([]byte, 44)
			r, e = channel.RW.Read(data)
			if r != 44 {
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
			fmt.Printf("Received (RAW) %x%x\n", header, data)
			fmt.Printf("Received %d: %x | %x\n", jobId, hash, odata)
		}
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
