package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type SerialChannel struct {
	Tag       string
	queryId   uint32
	Closed    bool
	RW        io.ReadWriteCloser
	mu        sync.Mutex
	callbacks map[uint32]chan []byte
}

func (channel *SerialChannel) PerformJob(data []byte, timeoutDuration int) ([]byte, error) {
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
	timeout := time.NewTimer(time.Second * time.Duration(timeoutDuration))
	go (func() {
		<-timeout.C

		channel.mu.Lock()
		ex := channel.callbacks[queryId]
		if ex != nil {
			delete(channel.callbacks, queryId)
		}
		channel.mu.Unlock()

		if ex != nil {
			ex <- make([]byte, 0)
		}
	})()

	// Sending
	// log.Printf("Sending [%v]: %d %x", channel.Tag, queryId, job)
	n, err := channel.RW.Write(job)
	if err != nil {
		return nil, err
	}
	if n != len(job) {
		return nil, errors.New("UART wirte issue")
	}

	r := <-ch
	if len(r) == 0 {
		return nil, errors.New("UART timed out")
	}
	return r, nil
}

func ReadAll(reader io.Reader, size int) ([]byte, error) {
	bt := make([]byte, 1)
	buffer := bytes.NewBuffer(nil)
	read := 0
	for {
		n, err := reader.Read(bt)
		switch err {
		case io.EOF:
			// log.Println("EOF")
			continue
		case nil:
		default:
			return nil, err
		}
		if n == 0 {
			// log.Println("EMPTY")
			continue
		}
		buffer.WriteByte(bt[0])
		// log.Printf("Received %x", bt[0])
		read++
		if read >= size {
			break
		}
	}
	return buffer.Bytes(), nil
}

func (channel *SerialChannel) Start() {
	go (func() {

		for {

			// Read header
			header, e := ReadAll(channel.RW, 4)
			if e != nil {
				log.Panic(e)
			}
			if header[0] != 0x42 || header[1] != 0x94 || header[2] != 0x37 || header[3] != 0x9b {
				log.Panicf("Invalid header %x", header)
			}

			// Read data
			data, e := ReadAll(channel.RW, 36+NonceSize)
			if e != nil {
				log.Panic(e)
			}

			// Parse packet
			jobId := binary.BigEndian.Uint32(data)
			body := data[4:]

			// Print result
			// log.Printf("Received (RAW) %x %x\n", header, data)
			// log.Printf("Received (BODY) %d: %x\n", jobId, body)

			// Response
			channel.mu.Lock()
			rch := channel.callbacks[jobId]
			if rch != nil {
				delete(channel.callbacks, jobId)
			}
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
		return &SerialChannel{RW: res, Closed: false, queryId: 0, callbacks: make(map[uint32]chan []byte), Tag: path}, nil
	}
}
