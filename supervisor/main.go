package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Version string `json:"version"`
	Url     string `json:"url"`
}

var client = &http.Client{Timeout: 10 * time.Second}

func doLoadConfig() (config *Config, err error) {
	resp, err := client.Get("https://pool.fra1.digitaloceanspaces.com/latest.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := new(Config)
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func loadConfig() Config {
	for {
		c, e := doLoadConfig()
		if e == nil {
			return *c
		}
		log.Println(e)
		time.Sleep(5 * time.Second)
	}
}

func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), f.Mode())
			if err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadPackage(config Config) {

	// Recreate tmp directory
	err := os.RemoveAll("/monad/imperium/software/tmp")
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("/monad/imperium/software/tmp", os.ModePerm)
	if err != nil {
		panic(err)
	}

	// Download package
	resp, err := http.Get(config.Url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(err)
	}

	// Create the file
	out, err := os.Create("/monad/imperium/software/tmp/output.zip")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	// Unzip
	err = Unzip("/monad/imperium/software/tmp/output.zip", "/monad/imperium/software/tmp/extracted")
	if err != nil {
		panic(err)
	}

	// Write descriptor
	cfg, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("/monad/imperium/software/tmp/extracted/config.json", cfg, 0644)
	if err != nil {
		panic(err)
	}
}

func applyPackage() error {
	err := os.RemoveAll("/monad/imperium/software/work/")
	if err != nil {
		return err
	}
	err = os.Rename("/monad/imperium/software/tmp/extracted", "/monad/imperium/software/work")
	if err != nil {
		return err
	}

	return nil
}

func stopAgent() {
	fmt.Println("Stopping agent...")
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9001/program/stop/agent", bytes.NewBuffer(make([]byte, 0)))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("Unable to stop agent")
	}
	time.Sleep(5 * time.Second)
	fmt.Println("Stopped agent")
}

func startAgent() {
	fmt.Println("Starting agent...")
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9001/program/start/agent", bytes.NewBuffer(make([]byte, 0)))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("Unable to stop agent")
	}
	time.Sleep(5 * time.Second)
	fmt.Println("Started agent")
}

func main() {

	// Loading current version
	currentVersion := "#invalid#"
	if _, err := os.Stat("/monad/imperium/software/work/config.json"); os.IsExist(err) {
		fmt.Println("Fetching latest config...")
	} else {
		fmt.Println("No existing package. Downloading package")
	}

	// Prepare package
	config := loadConfig()
	fmt.Printf("Found new version: %s\n", config.Version)
	if config.Version != currentVersion {
		// Download update
		fmt.Printf("Downloading %s\n", config.Version)
		downloadPackage(config)
		fmt.Println("Downloaded")

		// Stop Agent
		stopAgent()

		// Apply package
		applyPackage()
	}

	// Start
	startAgent()

	// Start refresh loop
	go (func() {
		for {
			nc := loadConfig()
			if nc.Version != config.Version {

				fmt.Printf("Found new version: %s\n", nc.Version)

				// Download
				downloadPackage(nc)

				// Stop
				stopAgent()

				// Apply package
				err := applyPackage()
				if err != nil {
					panic(err)
				}

				// Start
				startAgent()

				// Update config
				config = nc
			}
			time.Sleep(5 * time.Second)
		}
	})()

	select {}
}
