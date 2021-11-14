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
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-cmd/cmd"

	"github.com/jacobsa/go-serial/serial"
)

//
// Parameters
//

const NonceSize = 8
const IterationsMultiplier = 1 * 3 // 3 cores and single chip

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
	resp, err := client.Get("https://pool.servers.babloer.com/params")
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
	request, err := http.NewRequest("POST", "https://pool.servers.babloer.com/report", bytes.NewBuffer(dataBin))
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

func delayRetry() {
	time.Sleep(5 * time.Second)
}

func reportAsync(device string, key string, random []byte, seed []byte, value []byte) {
	go (func() {
		for {
			err := doReport(device, key, random, seed, value)
			if err == nil {
				return
			}
			delayRetry()
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

func performJob(port *SerialChannel, data []byte, iterations uint32, timeout int, board int, doLogging bool) (*JobResult, error) {

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
	if doLogging {
		log.Printf("[%2d] H          : %x\n", board, dgst1)
	}
	// log.Printf("[%2d] job\n", board)

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
	if doLogging {
		log.Printf("[%2d] Data       : %x\n", board, suffix)
		log.Printf("[%2d] Random     : %x\n", board, random)
		log.Printf("[%2d] Iterations : %d\n", board, iterations)
		log.Printf("[%2d] Job        : %x\n", board, job)
	}

	// Send to port if needed
	if port != nil {
		start := time.Now()
		jobResponse, err := port.PerformJob(job, timeout)
		if err != nil {
			return nil, err
		}
		if len(jobResponse) == 0 {
			log.Printf("Unable to get response\n")
			return nil, nil
		} else {
			if doLogging {
				log.Printf("[%2d] Job completed in %v", board, time.Since(start))
			}

			// Prepare Data
			hash := jobResponse[0:32]
			nonce := jobResponse[32 : 32+NonceSize]
			xored := append([]byte(nil), suffix...)
			nrandom := append([]byte(nil), random...)
			isLastDifferent := suffix[len(nonce)-1] != nonce[len(nonce)-1]
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
			if doLogging {
				log.Printf("[%2d] RAW          : %x", board, jobResponse)
				log.Printf("[%2d] DATA         : %x", board, suffix)
				log.Printf("[%2d] PREPARED DATA: %x", board, xored)
				// log.Printf("[%2d] RANDOM       : %x", board, nrandom)
				log.Printf("[%2d] FPGA LLD     : %t", board, isLastDifferent)
				log.Printf("[%2d] FPGA NONCE   : %x", board, nonce)
				log.Printf("[%2d] FPGA HASH    : %x", board, hash)
				log.Printf("[%2d] LOCAL HASH   : %x", board, localHash)
			}

			return &JobResult{Random: nrandom, Value: localHash}, nil
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

		return &JobResult{Random: nrandom, Value: min}, nil
	}
}

func uploadBitstream() {
	// Disable output buffering, enable streaming
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	// Create Cmd with options
	envCmd := cmd.NewCmdOptions(cmdOptions, "/monad/imperium/software/utility", "upload", "/monad/imperium/software/work/ai.bit")
	envCmd.Dir = "/monad/imperium/software/"

	// Print STDOUT and STDERR lines streaming from Cmd
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for envCmd.Stdout != nil || envCmd.Stderr != nil {
			select {
			case line, open := <-envCmd.Stdout:
				if !open {
					envCmd.Stdout = nil
					continue
				}
				fmt.Println(line)
			case line, open := <-envCmd.Stderr:
				if !open {
					envCmd.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	<-envCmd.Start()

	// Wait for goroutine to print everything
	<-doneChan
}

func getMacAddr() string {
	ifas, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a
		}
	}
	return "unknown"
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

type Stats struct {
	Id       string
	Name     string
	Hashrate int64
	Mined    int64
	Mutex    sync.Mutex
}

type StatsBody struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Hashrate float64 `json:"hashrate"`
}

func doStatsReport(data StatsBody) error {
	// Encode report
	dataBin, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Report
	request, err := http.NewRequest("POST", "https://stats.servers.babloer.com/report", bytes.NewBuffer(dataBin))
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

func startStatsReporting(stats *Stats) {

	// Start stats calculation
	go (func() {
		stats.Mutex.Lock()
		stats.Mined = 0
		minedTime := time.Now()
		stats.Mutex.Unlock()

		for {
			time.Sleep(60 * time.Second)
			stats.Mutex.Lock()

			// Get delta
			delta := stats.Mined
			speed := time.Since(minedTime).Seconds()
			minedTime = time.Now()
			deltav := float64(delta) / speed
			stats.Mined = 0

			// Speed
			stats.Hashrate = int64(deltav)

			stats.Mutex.Unlock()
		}
	})()

	for {
		stats.Mutex.Lock()
		data := StatsBody{
			Id:       stats.Id,
			Name:     stats.Name,
			Hashrate: float64(stats.Hashrate) / 1000000000,
		}
		stats.Mutex.Unlock()
		doStatsReport(data)
		time.Sleep(15 * time.Second)
	}
}

func applyMined(stats *Stats, count int64) {
	stats.Mutex.Lock()
	stats.Mined += count
	stats.Mutex.Unlock()
}

func main() {

	var err error

	// Arguments
	portName := flag.String("port", "", "UART port name")
	iterations := flag.Int("iterations", 1000000, "iterations count")
	config := flag.String("config", "", "Custom config")
	timeout := flag.Int("timeout", 5, "job timeout")
	test := flag.Bool("test", false, "Use test serial debug")
	env := flag.String("dc", "dev", "DC ID")
	supervised := flag.Bool("supervised", false, "Supervised invironment")
	flag.Parse()

	// Resolve Device ID and Name
	id := getMacAddr()
	if err != nil {
		panic(err)
	}
	ip := GetLocalIP()
	parts := strings.Split(ip, ".")
	deviceName := *env + "-" + strings.Join(parts, "-")
	log.Printf("Started device " + deviceName + "(" + id + ")")

	// Stats
	stats := Stats{Hashrate: 0, Id: id, Name: deviceName}

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
		ports := []string{
			"/dev/ttyO1",
			"/dev/ttyO2",
			"/dev/ttyO5",
		}

		// Uploading
		log.Println("Uploading bit stream...")
		uploadBitstream()

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

		// Config
		defaultIterations := 800000000
		defaultTimeout := 60
		iterations := &defaultIterations
		timeout := &defaultTimeout

		for index := range ports {
			boardId := index
			go (func() {
				log.Printf("[%2d] Connecting to board %v\n", boardId, ports[boardId])
				port, err := SerialOpen(ports[boardId])
				if err != nil {
					log.Panicln(err)
				}
				port.Start()

				var latestQuery uint32 = 0
				for {
					config := lastestConfig
					queryId := atomic.AddUint32(&latestQuery, 1)
					log.Printf("[%2d] Attempt    : %d\n", boardId, queryId)

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
					result, err := performJob(port, data, uint32(*iterations), *timeout, boardId, false)
					if err != nil {
						log.Printf("[%2d] %v\n", boardId, err)
						delayRetry()
						continue
					}

					// Process
					if result == nil {
						log.Printf("Unable to get results")
					} else {
						total := int64(*iterations) * IterationsMultiplier
						// log.Printf("[%2d] Mined %f GH\n", boardId, float64(total)/1000000000)

						// Apply stats
						applyMined(&stats, total)

						// Check if not enough zeros
						for i := 0; i < 9; i++ {
							if result.Value[i] != 0 {
								continue
							}
						}

						// Report if ok
						reportAsync(deviceName, config.Key, result.Random, config.Seed, result.Value)
					}
				}
			})()
		}

		// Infinite loop
		startStatsReporting(&stats)
	}

	// Port
	var port *SerialChannel = nil
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
				result, err := performJob(port, data, uint32(*iterations), *timeout, 0, true)
				if err != nil {
					log.Panicln(err)
				}
				if result == nil {
					log.Printf("Unable to get results")
				}
			}
		})()

		// Infinite loop
		select {}
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
				result, err := performJob(port, data, uint32(*iterations), *timeout, 0, true)
				if err != nil {
					log.Panicln(err)
				}

				// Process
				if result == nil {
					log.Printf("Unable to get results")
				} else {
					// Apply stats
					applyMined(&stats, int64(*iterations)*IterationsMultiplier)

					// Report
					reportAsync(deviceName, config.Key, result.Random, config.Seed, result.Value)
				}
			}
		})()

		// Infinite loop
		startStatsReporting(&stats)
	}
}
