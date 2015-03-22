// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

func readDeviceConfig(r io.Reader) (devices []*DeviceConfigEntry, err error) {
	reader := csv.NewReader(r)
	firstLine := true
	for {
		var record []string
		record, err = reader.Read()
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		}
		if firstLine {
			// skip header line
			firstLine = false
		} else {
			var device *DeviceConfigEntry
			device, err = NewDeviceConfigEntryFromCSV(record)
			if err != nil {
				return
			}
			if device.Dead != "" {
				continue
			}
			devices = append(devices, device)
		}
	}
	return
}

// Network device database, stored as CSVs.

type DeviceConfigEntry struct {
	IP, MAC, Name, Type, Kids, Dead, Owner, Description string
	SerialNumber                                        string
	ModelNumber                                         string
	AssetTag                                            string
}

func NewDeviceConfigEntryFromCSV(csv []string) (*DeviceConfigEntry, error) {
	if len(csv) != 11 {
		return nil, fmt.Errorf("Expected 11 fields:", csv)
	}
	return &DeviceConfigEntry{csv[0], csv[1], csv[2], csv[3], csv[4], csv[5], csv[6],
		csv[7], csv[8], csv[9], csv[10]}, nil
}
