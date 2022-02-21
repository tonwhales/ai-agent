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
)

//
// Parameters
//

const NonceSize = 8
const IterationsMultiplier = 1 * 3 // 4 cores per chip

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
	Device  string `json:"device"`
	Key     string `json:"key"`
	Random  string `json:"random"`
	Value   string `json:"value"`
	Expires uint32 `json:"expires"`
}

func doReport(device string, key string, random []byte, seed []byte, value []byte, expires uint32) error {

	// Encode report
	data := Report{
		Device:  device,
		Key:     key,
		Random:  base64.StdEncoding.EncodeToString(random),
		Value:   base64.StdEncoding.EncodeToString(value),
		Expires: expires,
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

func reportAsync(device string, key string, random []byte, seed []byte, value []byte, expires uint32) {
	go (func() {
		for {
			err := doReport(device, key, random, seed, value, expires)
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
	Expires uint32
	Random  []byte
	Value   []byte
}

func performJob(port *SerialChannel, data []byte, iterations uint32, timeout int, board int, chip int, doLogging bool) (*JobResult, error) {

	// Hash prefix
	expiresData := data[7:11]
	expires := binary.BigEndian.Uint32(expiresData)
	prefix1 := append([]byte(nil), data[:64]...)
	prefix2 := append([]byte(nil), data[:64]...)
	prefix3 := append([]byte(nil), data[:64]...)
	prefix4 := append([]byte(nil), data[:64]...)
	binary.BigEndian.PutUint32(prefix2[7:11], expires-1)
	binary.BigEndian.PutUint32(prefix3[7:11], expires-2)
	binary.BigEndian.PutUint32(prefix4[7:11], expires-3)

	dg := &digest{}
	dg.h[0] = init0
	dg.h[1] = init1
	dg.h[2] = init2
	dg.h[3] = init3
	dg.h[4] = init4
	dg.h[5] = init5
	dg.h[6] = init6
	dg.h[7] = init7
	blockGeneric(dg, prefix1)
	dgst1 := getDigest(*dg)

	dg = &digest{}
	dg.h[0] = init0
	dg.h[1] = init1
	dg.h[2] = init2
	dg.h[3] = init3
	dg.h[4] = init4
	dg.h[5] = init5
	dg.h[6] = init6
	dg.h[7] = init7
	blockGeneric(dg, prefix2)
	dgst2 := getDigest(*dg)

	dg = &digest{}
	dg.h[0] = init0
	dg.h[1] = init1
	dg.h[2] = init2
	dg.h[3] = init3
	dg.h[4] = init4
	dg.h[5] = init5
	dg.h[6] = init6
	dg.h[7] = init7
	blockGeneric(dg, prefix3)
	dgst3 := getDigest(*dg)

	dg = &digest{}
	dg.h[0] = init0
	dg.h[1] = init1
	dg.h[2] = init2
	dg.h[3] = init3
	dg.h[4] = init4
	dg.h[5] = init5
	dg.h[6] = init6
	dg.h[7] = init7
	blockGeneric(dg, prefix4)
	dgst4 := getDigest(*dg)

	if doLogging {
		log.Printf("[%2d] H          : %x,%x,%x,%x\n", board, dgst1, dgst2, dgst3, dgst4)
	}

	// Suffix
	suffix := data[64:]
	random := suffix[27:]
	suffix = append(suffix, 0x80, 0x00, 0x00, 0x00, 0x00)

	// Calculate job
	job := []byte{}
	job = append(job, dgst1...)
	job = append(job, dgst2...)
	job = append(job, dgst3...)
	job = append(job, dgst4...)
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
		jobResponse, err := port.PerformJob(chip, job, timeout)
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
			nonce := jobResponse[32 : 32+4]
			prefixIdRaw := jobResponse[32+4 : 32+4+4]
			prefixId := binary.BigEndian.Uint32(prefixIdRaw)
			xored := append([]byte(nil), suffix...)
			nrandom := append([]byte(nil), random...)
			index := nonce[len(nonce)-1] - suffix[len(nonce)-1]
			for i := 0; i < len(nonce); i++ {
				xored[i] = nonce[i]
				xored[i+48] = nonce[i]
				nrandom[i+21] = nonce[i]
			}
			log.Printf("[%2d] PREFIX ID          : %d", board, prefixId)

			// TODO:L
			resExpires := expires
			prefix := prefix1
			if prefixId == 2 {
				prefix = prefix2
				resExpires = resExpires - 1
			}
			if prefixId == 3 {
				prefix = prefix3
				resExpires = resExpires - 2
			}
			if prefixId == 4 {
				prefix = prefix4
				resExpires = resExpires - 3
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
				log.Printf("[%2d] FPGA LLD     : %d", board, index)
				log.Printf("[%2d] FPGA NONCE   : %x", board, nonce)
				log.Printf("[%2d] FPGA HASH    : %x", board, hash)
				log.Printf("[%2d] LOCAL HASH   : %x", board, localHash)
			}

			// Check hash
			if !bytes.Equal(hash, localHash) {
				return nil, fmt.Errorf("hash mismatch. Expected %x, but got %x", localHash, hash)
			}

			return &JobResult{Random: nrandom, Value: localHash, Expires: resExpires}, nil
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
			sh.Write(data)
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

func uploadBitstream(name string) {
	// Disable output buffering, enable streaming
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	// Create Cmd with options
	envCmd := cmd.NewCmdOptions(cmdOptions, "/monad/imperium/software/utility", "upload", "/monad/imperium/software/work/"+name)
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
	Id           string
	Name         string
	Datacenter   string
	Hashrate     int64
	Mined        int64
	Mutex        sync.Mutex
	Temperatures [][]float32
}

type StatsBody struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Datacenter   string            `json:"dc"`
	Hashrate     float64           `json:"hashrate"`
	Temperatures []TemperatureBody `json:"temperature"`
}
type TemperatureBody struct {
	Id    string  `json:"id"`
	Value float32 `json:"value"`
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
		temperatures := make([]TemperatureBody, 0)
		for board := 0; board < 3; board++ {
			for chip := 0; chip < 6; chip++ {
				temperatures = append(temperatures, TemperatureBody{
					Id:    fmt.Sprintf("chip_%d_%d", board, chip),
					Value: stats.Temperatures[board][chip],
				})
			}
		}
		data := StatsBody{
			Id:           stats.Id,
			Name:         stats.Name,
			Datacenter:   stats.Datacenter,
			Hashrate:     float64(stats.Hashrate) / 1000000000,
			Temperatures: temperatures,
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
	chip := flag.Int("chip", 6, "Working Chip ID")
	bitstream := flag.String("bitstream", "ai.bit", "Bitstream to use")
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
	stats := Stats{Hashrate: 0, Id: id, Name: deviceName, Datacenter: *env, Temperatures: [][]float32{{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}}}

	// Test
	if test != nil && *test {
		if portName == nil || *portName == "" {
			log.Panicln("No port specified")
		}

		log.Println("Connecting to COM port...")
		pp, err := SerialOpen(*portName, 115200)
		if err != nil {
			log.Panicln(err)
		}

		// Read cycle
		// go (func() {
		// 	for {
		// 		frame, err := pp.Read()
		// 		if err != nil {
		// 			log.Panicln(err)
		// 		}
		// 		log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)
		// 	}
		// })()

		// Write data
		log.Println("Wait...")

		// Fastclock
		// frame, err := pp.Request(*chip, 0xA2, []byte{0x0F, 0x01})
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)

		// // Set frequency
		// err = pp.SetFrequency(*chip, 100)
		// if err != nil {
		// 	log.Panicln(err)
		// }

		// // CoinID
		// frame, err = pp.Request(*chip, 0xA2, []byte{0x20})
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)

		// // Cores
		// frame, err = pp.Request(*chip, 0xA2, []byte{0x21})
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)

		// Write job
		// job, err := hex.DecodeString("de557f05c301618d5f527f856bff0452611101e9e601a81c0fbba95cbc0eb3fdaf50c54acefc0817ac0465fe3df51f0671341bfca85a552fc1b1f3adb73d4108996e7466b32669a0a174eb397ddee4c9af50c54acefc0817ac046580000000002faf0800")
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// frame, err = pp.Request(*chip, 0xA2, append([]byte{0x11, 0x00}, job...))
		// if err != nil {
		// 	log.Panicln(err)
		// }
		// log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)

		// on := true
		// for {
		// 	time.Sleep(1 * time.Second)
		// 	if on {
		// 		str := "8c 00 00 00 00 fa 8c c0 0a 18 ed 5b e1 a9 4c c5 62 7e 2b 11 b7 ed 25 3a b0 c9 9a 98 d8 e0 12 06 86 1b 02 cf 4d bb 3a 00 9a 01 50 cf 88 f5 c8 9e 47 ea 0a 1b 03 0d 5e 23 c5 35 de 8a c4 d9 e2 b4 2c e7 28 ed 16 91 ca 9e 25 45 ef f3 f8 93 a5 3f 80 5c 69 e7 5d 8d 3e 3a 00 9a 01 50 cf 88 f5 c8 9e 47 80 00 00 00 00 00 0f 42 40"
		// 		str = strings.ReplaceAll(str, " ", "")
		// 		data, err := hex.DecodeString(str)
		// 		if err != nil {
		// 			log.Panicln(err)
		// 		}
		// 		err = pp.Write(*chip, 0x0, data)
		// 		if err != nil {
		// 			log.Panicln(err)
		// 		}
		// 		log.Printf("Written %x\n", data)
		// 	} else {
		// 		data, err := hex.DecodeString("9A")
		// 		if err != nil {
		// 			log.Panicln(err)
		// 		}
		// 		err = pp.Write(*chip, 0x0, data)
		// 		if err != nil {
		// 			log.Panicln(err)
		// 		}
		// 		log.Printf("Written %x\n", data)

		// 		// 0200000600019a738103
		// 		// 0200010600019ab3bc03
		// 	}
		// 	on = !on
		// }

		for {
			time.Sleep(1 * time.Second)

			// Temp
			pp.Write(*chip, 0x00, []byte{0x7c, 0x0e, 0x00, 0x00, 0x00})
			frame, err := pp.Read()
			if err != nil {
				log.Panicln(err)
			}
			log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)
			x := float32(binary.BigEndian.Uint16(frame.Data[1:]))
			temp := x*502.9098/65536 - 273.819
			log.Printf("Temperature %f", temp)

			// Status
			// pp.Write(*chip, 0xA2, []byte{0x12, 0x00})
			// frame, err = pp.Read()
			// if err != nil {
			// 	log.Panicln(err)
			// }
			// log.Printf("Frame from %d: %x \n", frame.ChipID, frame.Data)
		}

		// bt := make([]byte, 1)
		// for {
		// 	log.Println("Reading...")
		// 	r, err := pp.Read(bt)
		// 	if err != nil {
		// 		switch err {
		// 		case io.EOF:
		// 			continue
		// 		default:
		// 			log.Panic(err)
		// 		}
		// 	}
		// 	if r != 1 {
		// 		log.Panic("Empty")
		// 	}
		// 	log.Printf("%x", bt)
		// }
	}

	// Check supervised flag
	if supervised != nil && *supervised {
		log.Println("Running in supervised mode")

		// Start Leds
		StartLed()

		// Define Ports
		ports := []string{
			"/dev/ttyO1",
			"/dev/ttyO2",
			"/dev/ttyO5",
		}
		chips := []int{
			1,
			2,
			3,
			4,
			5,
			6,
		}

		// Uploading
		// SetGreenLed(true, true) // It seems that uploadBitstream enables green led blinking anyway
		SetRedLed(false, false)
		log.Println("Uploading bit stream...")
		uploadBitstream(*bitstream)
		SetGreenLed(true, true)

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
				log.Printf("[%2d] Connecting to board\n", boardId)
				port, err := SerialOpen(ports[boardId], 115200)
				if err != nil {
					log.Panicln(err)
				}

				log.Printf("[%2d] Starting threads\n", boardId)
				var latestQuery uint32 = 0
				for chipIndex := range chips {
					chipId := chips[chipIndex]

					// Jobs
					go (func() {
					outer:
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
							result, err := performJob(port, data, uint32(*iterations), *timeout, boardId, chipId, false)
							if err != nil {
								log.Printf("[%2d] %v\n", boardId, err)
								delayRetry()
								continue
							}
							if result == nil {
								log.Printf("Unable to get results")
								delayRetry()
								continue
							}

							// Apply stats
							applyMined(&stats, int64(*iterations)*IterationsMultiplier)

							// Check if not enough zeros
							for i := 0; i < 4; i++ {
								if result.Value[i] != 0 {
									continue outer
								}
							}
							if result.Value[4] > 0x0f {
								continue outer
							}

							// Report
							reportAsync(deviceName, config.Key, result.Random, config.Seed, result.Value, result.Expires)
						}
					})()

					// Monitoring
					go func() {
						for {

							// Collect temperature
							v, err := port.GetTemperature(chipId)
							if err != nil {
								log.Printf("[%2d] %v\n", boardId, err)
								delayRetry()
								continue
							}
							stats.Temperatures[boardId][chipId-1] = v

							// Delay
							delayRetry()
						}
					}()
				}
			})()
		}

		go (func() {
			time.Sleep(20 * time.Second)

			for {
				// Monitor hashrate
				if stats.Hashrate < 1000 {
					SetRedLed(true, true)
					SetGreenLed(false, false)
				} else {
					SetRedLed(false, false)
					SetGreenLed(true, true)
				}

				// Delay
				delayRetry()
			}
		})()

		// Infinite loop
		startStatsReporting(&stats)
	}

	// Port
	var port *SerialChannel = nil
	if portName != nil && *portName != "" {
		log.Println("Connecting to COM port...")
		port, err = SerialOpen(*portName, 115200)
		if err != nil {
			log.Panicln(err)
		}
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
				result, err := performJob(port, data, uint32(*iterations), *timeout, 0, *chip, true)
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
				result, err := performJob(port, data, uint32(*iterations), *timeout, 0, *chip, true)
				if err != nil {
					log.Println(err)
					continue
				}

				// Process
				if result == nil {
					log.Printf("Unable to get results")
				} else {
					// Apply stats
					applyMined(&stats, int64(*iterations)*IterationsMultiplier)

					// Report
					reportAsync(deviceName, config.Key, result.Random, config.Seed, result.Value, result.Expires)
				}
			}
		})()

		// Infinite loop
		startStatsReporting(&stats)
	}
}
