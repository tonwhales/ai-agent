package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type Config struct {
	ID        string
	Reference string
	Header    []byte
	Seed      []byte
}

var client = &http.Client{Timeout: 10 * time.Second}

type ApiConfig struct {
	Id     string `json:"id"`
	Ref    string `json:"ref"`
	Header string `json:"header"`
	Seed   string `json:"seed"`
}

func loadConfig() (config *Config, err error) {
	resp, err := client.Get("http://64.225.102.108:3000/params")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := new(ApiConfig)
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	header, err := base64.StdEncoding.DecodeString(res.Header)
	if err != nil {
		return nil, err
	}
	seed, err := base64.StdEncoding.DecodeString(res.Seed)
	if err != nil {
		return nil, err
	}
	r := Config{}
	r.ID = res.Id
	r.Reference = res.Ref
	r.Header = header
	r.Seed = seed
	return &r, nil
}

func loadConfigRetry() Config {
	for {
		c, e := loadConfig()
		if e == nil {
			return *c
		}
		time.Sleep(5 * time.Second)
	}
}

type Job struct {
	Data   []byte
	Random []byte
	Seed   []byte
}

func getDigest(d digest) []byte {
	var res = make([]byte, 32)
	binary.BigEndian.PutUint32(res[0:], d.h[0])
	binary.BigEndian.PutUint32(res[4:], d.h[1])
	binary.BigEndian.PutUint32(res[8:], d.h[2])
	binary.BigEndian.PutUint32(res[12:], d.h[3])
	binary.BigEndian.PutUint32(res[16:], d.h[4])
	binary.BigEndian.PutUint32(res[20:], d.h[5])
	binary.BigEndian.PutUint32(res[24:], d.h[6])
	binary.BigEndian.PutUint32(res[28:], d.h[7])
	return res
}

func SerialOpen(path string) (io.ReadWriteCloser, error) {
	return serial.Open(serial.OpenOptions{
		PortName: path,
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 2,
		// mode
		InterCharacterTimeout: 100,
		MinimumReadSize:       0,

		RTSCTSFlowControl: false,
	})
}

func main() {

	var port io.ReadWriteCloser = nil
	var err error

	// Arguments
	portName := flag.String("port", "", "UART port name")
	flag.Parse()
	if portName != nil && *portName != "" {
		fmt.Println("Connecting to COM port...")
		port, err = SerialOpen(*portName)
		if err != nil {
			log.Panicln(err)
		}
	} else {
		fmt.Println("Running without COM port")
	}

	// Loading config
	fmt.Println("Loading initial config...")
	lastestConfig := loadConfigRetry()

	// Start config refetch loop
	fmt.Println("Starting config refresh...")
	go (func() {
		for {
			lastestConfig = loadConfigRetry()
			time.Sleep(5 * time.Second)
		}
	})()

	// Start threads
	fmt.Println("Starting threads...")
	var latestQuery uint32 = 0
	go (func() {
		for {
			config := lastestConfig
			queryId := atomic.AddUint32(&latestQuery, 1)
			fmt.Printf("Attempt    : %d\n", queryId)

			// Create random
			random := make([]byte, 32)
			rand.Read(random)

			// Create block
			data := make([]byte, 0)
			data = append(data, 0x0, 0xF2)
			data = append(data, config.Header...)
			data = append(data, random...)
			data = append(data, config.Seed...)
			data = append(data, random...)

			// Hash prefix
			prefix := data[:64]
			dg := &digest{}
			dg.h[0] = init0
			dg.h[1] = init1
			dg.h[2] = init2
			dg.h[3] = init3
			dg.h[4] = init4
			dg.h[5] = init5
			dg.h[6] = init6
			dg.h[7] = init7
			blockGeneric(dg, prefix)
			dgst1 := getDigest(*dg)
			fmt.Printf("H          : %x\n", dgst1)

			// Suffix
			suffix := data[64:]
			suffix = append(suffix, 0x80, 0x00, 0x00, 0x00, 0x00)

			// Calculate job
			iterations := 1000000
			job := []byte{0xa4, 0x61, 0xa1, 0x8c}
			tmp := make([]byte, 4)
			binary.BigEndian.PutUint32(tmp, queryId)
			job = append(job, tmp...)
			job = append(job, dgst1...)
			job = append(job, suffix...)
			binary.BigEndian.PutUint32(tmp, uint32(iterations))
			job = append(job, tmp...)

			fmt.Printf("Data       : %x\n", suffix)
			fmt.Printf("Iterations : %d\n", iterations)
			fmt.Printf("Job        : %x\n", job)

			if port != nil {
				n, err := (port).Write(job)
				if err != nil {
					log.Panicln(err)
				}
				if n != len(job) {
					log.Panicln("Invalid bytes written")
				}
			}

			// Debug counter
			// counter := []byte{
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0, 0,
			// 	0, 0, 0, 0, 0, 0, 0x03, 0xD8}
			// blockGeneric(dg, suffix)
			// blockGeneric(dg, counter)
			// dgst := getDigest(*dg)
			// h := sha256.New()
			// h.Write(data)
			// fmt.Printf("%x\n", h.Sum(nil))
			// fmt.Printf("%x\n", dgst)

			// Delay
			time.Sleep(5 * time.Second)
		}
	})()

	// Infinite loop
	select {}
}
