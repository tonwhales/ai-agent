package main

import (
	"encoding/binary"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type SerialChannel struct {
	queryId   uint32
	Closed    bool
	RW        io.ReadWriteCloser
	mu        sync.Mutex
	callbacks map[uint32]chan []byte
}

func (channel *SerialChannel) PerformJob(data []byte) []byte {
	queryId := channel.queryId
	channel.queryId++

	// Package
	job := []byte{0xa4, 0x61, 0xa1, 0x8c}
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, queryId)
	job = append(job, tmp...)
	job = append(job, data...)

	// Persist job
	channel.mu.Lock()
	ch := make(chan []byte)
	channel.callbacks[queryId] = ch
	channel.mu.Unlock()

	// Timeout
	timeout := time.NewTimer(time.Second)
	go (func() {
		<-timeout.C

		channel.mu.Lock()
		ex := channel.callbacks[queryId]
		channel.callbacks[queryId] = nil
		channel.mu.Unlock()

		if ex != nil {
			ex <- make([]byte, 0)
		}
	})()

	// Sending
	n, err := channel.RW.Write(job)
	if err != nil {
		log.Panicln(err)
	}
	if n != len(job) {
		log.Panicln("Invalid write")
	}

	// Log result
	log.Printf("Sent job %d| %x\n", queryId, job)

	r := <-ch
	if len(r) == 0 {
		log.Panicln("Invalid response")
	}
	return r
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
			body := data[4:]

			// Print result
			log.Printf("Received (RAW) %x %x\n", header, data)
			log.Printf("Received (BODY) %d: %x\n", jobId, body)

			// Response
			channel.mu.Lock()
			rch := channel.callbacks[jobId]
			channel.callbacks[jobId] = nil
			channel.mu.Unlock()
			if rch != nil {
				rch <- body
			} else {
				log.Printf("Unable to find callback for %d\n", jobId)
			}
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
