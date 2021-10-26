package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sync/atomic"
	"time"
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

func main() {

	var port *SerialChannel = nil
	var err error

	// Arguments
	portName := flag.String("port", "", "UART port name")
	iterations := flag.Int("iterations", 1000000, "iterations count")
	flag.Parse()
	if portName != nil && *portName != "" {
		log.Println("Connecting to COM port...")
		port, err = SerialOpen(*portName)
		port.Start()
		if err != nil {
			log.Panicln(err)
		}
	} else {
		log.Println("Running without COM port")
	}

	// Loading config
	log.Println("Loading initial config...")
	lastestConfig := loadConfigRetry()

	// Start config refetch loop
	log.Println("Starting config refresh...")
	go (func() {
		for {
			lastestConfig = loadConfigRetry()
			time.Sleep(5 * time.Second)
		}
	})()

	// Start threads
	log.Println("Starting threads...")
	var latestQuery uint32 = 0
	go (func() {
		for {
			config := lastestConfig
			queryId := atomic.AddUint32(&latestQuery, 1)
			log.Printf("Attempt    : %d\n", queryId)

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
			log.Printf("H          : %x\n", dgst1)

			// Suffix
			suffix := data[64:]
			suffix = append(suffix, 0x80, 0x00, 0x00, 0x00, 0x00)

			// Calculate job
			job := []byte{}
			job = append(job, dgst1...)
			job = append(job, suffix...)
			tmp := make([]byte, 4)
			binary.BigEndian.PutUint32(tmp, uint32(*iterations))
			job = append(job, tmp...)

			// Display job
			log.Printf("Data       : %x\n", suffix)
			log.Printf("Iterations : %d\n", *iterations)
			log.Printf("Job        : %x\n", job)

			// Send to port if needed
			if port != nil {
				start := time.Now()
				jobResponse := port.PerformJob(job)
				if len(jobResponse) == 0 {
					log.Panicln("Unable to get response")
				} else {
					log.Printf("Job completed in %v", time.Since(start))

					// Prepare Data
					hash := jobResponse[0:32]
					nonce := jobResponse[32 : 32+4]
					xored := append([]byte(nil), suffix...)
					for i := 0; i < len(nonce); i++ {
						xored[i] = xored[i] ^ nonce[i]
						xored[i+11] = xored[i+11] ^ nonce[i]
					}

					// Check hash
					sh := sha256.New()
					sh.Write(prefix)
					sh.Write(suffix[:64-5])
					localHash := sh.Sum(nil)

					// Print results
					log.Printf("RAW       : %x", jobResponse)
					log.Printf("DATA      : %x", suffix)
					log.Printf("XORED DATA: %x", xored)
					log.Printf("FPGA HASH : %x", hash)
					log.Printf("LOCAL HASH: %x", localHash)
				}
			}

			// Delay
			time.Sleep(5 * time.Second)
		}
	})()

	// Infinite loop
	select {}
}
