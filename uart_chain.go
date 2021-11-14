package main

import (
	"encoding/binary"
	"errors"
	"time"
)

func (channel *SerialChannel) PerformJob(chipId int, data []byte, timeoutDuration int) ([]byte, error) {

	// Job ID
	queryId := channel.queryId
	channel.queryId++

	// Package
	job := []byte{0x8c}
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, queryId)
	job = append(job, tmp...)
	job = append(job, data...)
	channel.Write(chipId, job)

	// Check job
	start := time.Now()
	for {

		// Job timeout
		if time.Since(start).Seconds() >= float64(timeoutDuration) {
			return nil, errors.New("job timeout")
		}

		// Retry every 200 ms
		time.Sleep(200 * time.Millisecond)

		// Do check
		statusCheck := []byte{0x9a}
		res, err := channel.Request(chipId, statusCheck)
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
			return nil, errors.New("job mismatch")
		}

		// Job not ready
		if jobState == 0 {
			continue
		}

		// Job ready
		if jobState == 1 {
			return data, nil
		}

		// Invalid state
		return nil, errors.New("invalid job state")
	}
}
