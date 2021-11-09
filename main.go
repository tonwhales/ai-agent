package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type Config struct {
	Key    string
	Header []byte
	Seed   []byte
}

var client = &http.Client{Timeout: 10 * time.Second}

type ApiConfig struct {
	Key    string `json:"key"`
	Header string `json:"header"`
	Seed   string `json:"seed"`
}

func loadConfig() (config *Config, err error) {
	resp, err := client.Get("https://pool.tonwhales.com/params")
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
	r.Key = res.Key
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

type Report struct {
	Device string `json:"device"`
	Key    string `json:"key"`
	Random string `json:"random"`
	Value  string `json:"value"`
}

func doReport(device string, key string, random []byte, seed []byte, value []byte) error {

	// Encode report
	data := Report{
		Device: device,
		Key:    key,
		Random: base64.StdEncoding.EncodeToString(random),
		Value:  base64.StdEncoding.EncodeToString(value),
	}
	dataBin, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Report
	request, err := http.NewRequest("POST", "https://pool.tonwhales.com/report", bytes.NewBuffer(dataBin))
	if err != nil {
		fmt.Println(err)
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func reportAsync(device string, key string, random []byte, seed []byte, value []byte) {
	go (func() {
		for {
			err := doReport(device, key, random, seed, value)
			if err == nil {
				return
			}
			time.Sleep(5 * time.Second)
		}
	})()
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

type JobResult struct {
	Random []byte
	Value  []byte
}

func performJob(port *SerialChannel, data []byte, iterations uint32, timeout int) *JobResult {

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
	random := suffix[27:]
	suffix = append(suffix, 0x80, 0x00, 0x00, 0x00, 0x00)

	// Calculate job
	job := []byte{}
	job = append(job, dgst1...)
	job = append(job, suffix...)
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, iterations)
	job = append(job, tmp...)

	// Display job
	log.Printf("Data       : %x\n", suffix)
	log.Printf("Random     : %x\n", random)
	log.Printf("Iterations : %d\n", iterations)
	log.Printf("Job        : %x\n", job)

	// Send to port if needed
	if port != nil {
		start := time.Now()
		jobResponse := port.PerformJob(job, timeout)
		if len(jobResponse) == 0 {
			log.Printf("Unable to get response\n")
			return nil
		} else {
			log.Printf("Job completed in %v", time.Since(start))

			// Prepare Data
			hash := jobResponse[0:32]
			nonce := jobResponse[32 : 32+NonceSize]
			xored := append([]byte(nil), suffix...)
			nrandom := append([]byte(nil), random...)
			for i := 0; i < len(nonce); i++ {
				xored[i] = nonce[i]
				xored[i+48] = nonce[i]
				nrandom[i+21] = nonce[i]
			}

			// Check hash
			sh := sha256.New()
			sh.Write(prefix)
			sh.Write(xored[:64-5])
			localHash := sh.Sum(nil)

			// Print results
			log.Printf("RAW          : %x", jobResponse)
			log.Printf("DATA         : %x", suffix)
			log.Printf("PREPARED DATA: %x", xored)
			log.Printf("RANDOM       : %x", nrandom)
			log.Printf("FPGA NONCE   : %x", nonce)
			log.Printf("FPGA HASH    : %x", hash)
			log.Printf("LOCAL HASH   : %x", localHash)

			return &JobResult{Random: nrandom, Value: localHash}
		}
	} else {

		//
		// CPU-based miner
		//

		min := make([]byte, 32)
		minNonce := make([]byte, 4)
		for i := 0; i < int(iterations); i++ {

			// Create nonce
			nonce := make([]byte, 4)
			binary.BigEndian.PutUint32(nonce, uint32(i))
			xored := append([]byte(nil), suffix...)
			for i := 0; i < len(nonce); i++ {
				xored[i] = nonce[i]
				xored[i+48] = nonce[i]
			}

			// Calculate hash
			sh := sha256.New()
			sh.Write(prefix)
			sh.Write(xored[:64-5])
			localHash := sh.Sum(nil)

			// Update minimum
			if i == 0 {
				min = localHash
				minNonce = nonce
			} else {
				if bytes.Compare(localHash, min) < 0 {
					min = localHash
					minNonce = nonce
				}
			}
		}

		nrandom := append([]byte(nil), random...)
		for i := 0; i < len(minNonce); i++ {
			nrandom[i+21] = minNonce[i]
		}

		log.Printf("CPU RANDOM : %x", nrandom)
		log.Printf("CPU NONCE  : %x", minNonce)
		log.Printf("CPU HASH   : %x", min)

		return &JobResult{Random: nrandom, Value: min}
	}
}

func main() {

	var port *SerialChannel = nil
	var err error

	// Arguments
	portName := flag.String("port", "", "UART port name")
	iterations := flag.Int("iterations", 1000000, "iterations count")
	config := flag.String("config", "", "Custom config")
	timeout := flag.Int("timeout", 5, "job timeout")
	deviceName := flag.String("name", "dev", "Device name")
	test := flag.Bool("test", false, "Use test serial debug")
	supervised := flag.Bool("supervised", false, "Supervised invironment")
	flag.Parse()

	// Test
	if test != nil && *test {
		if portName == nil || *portName == "" {
			log.Panicln("No port specified")
		}

		log.Println("Connecting to COM port...")
		pp, err := serial.Open(serial.OpenOptions{
			PortName: *portName,
			BaudRate: 115200,
			DataBits: 8,
			StopBits: 2,

			// mode
			InterCharacterTimeout: 100,
			MinimumReadSize:       0,

			RTSCTSFlowControl: false,
		})
		if err != nil {
			log.Panicln(err)
		}

		bt := make([]byte, 1)
		for {
			log.Println("Reading...")
			r, err := pp.Read(bt)
			if err != nil {
				switch err {
				case io.EOF:
					continue
				default:
					log.Panic(err)
				}
			}
			if r != 1 {
				log.Panic("Empty")
			}
			log.Printf("%x", bt)
		}
	}

	// Check supervised flag
	if supervised != nil && *supervised {
		log.Println("Running in supervised mode")
	}

	// Port
	if portName != nil && *portName != "" {
		log.Println("Connecting to COM port...")
		port, err = SerialOpen(*portName)
		if err != nil {
			log.Panicln(err)
		}
		port.Start()
	} else {
		log.Println("Running without COM port")
	}

	// Debug mode
	if config != nil && *config != "" {
		// Loading config
		log.Println("Loading TEST config...")
		hextData, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Panicln(err)
		}
		if len(hextData) != 246 {
			log.Fatalf("Invalid file length. Expected 246, got %d", len(hextData))
		}
		data := make([]byte, 123)
		l, err := hex.Decode(data, hextData)
		if err != nil {
			log.Panicln(err)
		}
		if l != 123 {
			log.Fatalf("Invalid hex length. Expected 123, got %d", l)
		}

		// Start
		var latestQuery uint32 = 0
		go (func() {
			for {
				queryId := atomic.AddUint32(&latestQuery, 1)
				log.Printf("Attempt    : %d\n", queryId)

				// Do Job
				result := performJob(port, data, uint32(*iterations), *timeout)
				if result == nil {
					log.Printf("Unable to get results")
				}
			}
		})()
	} else {

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
				data = append(data, config.Header...)
				data = append(data, random...)
				data = append(data, config.Seed...)
				data = append(data, random...)

				// Do Job
				result := performJob(port, data, uint32(*iterations), *timeout)

				// Process
				if result == nil {
					log.Printf("Unable to get results")
				} else {
					reportAsync(*deviceName, config.Key, result.Random, config.Seed, result.Value)
				}
			}
		})()
	}

	// Infinite loop
	select {}
}
