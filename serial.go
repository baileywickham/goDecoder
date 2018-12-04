package main

import (
	"errors"
	"go.bug.st/serial.v1"
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

func closePort() error {
	err := currentPort.Close()
	return err
}
