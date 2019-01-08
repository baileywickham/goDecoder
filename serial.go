package main

import (
	"errors"
	"go.bug.st/serial.v1"
	"time"
)

func listPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	//how do I handle errors
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, errors.New("No Ports Returned")
	}

	return ports, nil

	// This should never happen
}

var currentPort serial.Port

func connectPort(p string) error {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	// Creates global port object
	var err error
	currentPort, err = serial.Open(p, mode)
	return err
}

func writePort(data []byte) error {
	// Should we handle the num of bits written?
	_, err := currentPort.Write(data)
	return err
}

func closePort() {
	// Allows safe closing of ports
	if &currentPort == nil {
		return
	}
	if err := currentPort.Close(); err != nil {
		panic(err)
	}

}

func readPort(data chan<- Response) error {
	for {
		// Constantly reads and waits for data, then returns
		// data into a channel. this should make reading sort
		// of async
		buff := make([]byte, 1024)
		n, err := currentPort.Read(buff)
		if err != nil {
			return err
		}
		if n == 0 {
			time.Sleep(100 * time.Millisecond)
		} else {
			data <- Response{true, "", nil, buff[:n]}
		}
	}
}
