package main

import (
	"encoding/binary"
)

func (channel *SerialChannel) PerformJob(chipId int, data []byte, timeoutDuration int) ([]byte, error) {
	queryId := channel.queryId
	channel.queryId++

	// Package
	job := []byte{0xa4, 0x61, 0xa1, 0x8c}
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, queryId)
	job = append(job, tmp...)
	job = append(job, data...)

	// Persist job
	channel.Write(chipId, job)
	res, err := channel.Read()
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}
