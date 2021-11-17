package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

func (channel *SerialChannel) PerformJob(chipId int, data []byte, timeoutDuration int) ([]byte, error) {

	// Job ID
	queryId := channel.queryId
	channel.queryId++
	statusCheck := []byte{0x9a}

	// Preflight check
	// res, err := channel.Request(chipId, 0x0, statusCheck)
	// if err != nil {
	// 	return nil, err
	// }
	// if len(res.Data) == 0 {
	// 	return nil, errors.New("invalid frame")
	// }

	// Package
	job := []byte{0x8c}
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, queryId)
	job = append(job, tmp...)
	job = append(job, data...)
	err := channel.Write(chipId, 0x00, job)
	if err != nil {
		return nil, err
	}

	// Check job
	start := time.Now()
	for {

		// Job timeout
		if time.Since(start).Seconds() >= float64(timeoutDuration) {
			return nil, errors.New("job timeout")
		}

		// Retry every 200 ms
		time.Sleep(100 * time.Millisecond)

		// Do check
		res, err := channel.Request(chipId, 0x0, statusCheck)
		if err != nil {
			return nil, err
		}
		if len(res.Data) == 0 {
			return nil, errors.New("invalid frame")
		}

		// No job
		jobState := res.Data[0]
		data := res.Data[1:]
		if jobState == 0 {
			return nil, errors.New("no job found")
		}

		// Parse package
		receivedJobId := binary.BigEndian.Uint32(data)
		data = data[4:]

		// Check job id
		if receivedJobId != queryId {
			return nil, fmt.Errorf("job mismatch. expected: %x, got: %x", queryId, receivedJobId)
		}

		// Job not ready
		if jobState == 1 {
			continue
		}

		// Job ready
		if jobState == 2 {
			return data, nil
		}

		// Invalid state
		return nil, errors.New("invalid job state")
	}
}
