// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package db

import (
	"net"
	"time"
)

type DeviceIP net.IP

func (d DeviceIP) Equal(x DeviceIP) bool {
	return net.IP(d).Equal(net.IP(x))
}

func (d DeviceIP) String() string {
	return net.IP(d).String()
}

func (d DeviceIP) MarshalText() ([]byte, error) {
	return net.IP(d).MarshalText()
}

func (d *DeviceIP) UnmarshalText(text []byte) error {
	return (*net.IP)(d).UnmarshalText(text)
}

func ParseDeviceIP(s string) DeviceIP {
	return DeviceIP(net.ParseIP(s))
}

type Device struct {
	IP          DeviceIP
	Name        string
	ActiveUntil time.Time
}

func NewDevice(ip string, name string) Device {
	return Device{ParseDeviceIP(ip), name, time.Time{}}
}

// Helper func for modifying activeUntil
func modifyActiveUntil(oldActiveUntil time.Time, delta time.Duration,
	baseTime time.Time) (newActiveUntil time.Time) {
	if oldActiveUntil.Before(baseTime) {
		oldActiveUntil = baseTime
	}
	newActiveUntil = oldActiveUntil.Add(delta)
	if newActiveUntil.Before(baseTime) {
		newActiveUntil = baseTime
	}
	return
}

type DB interface {
	Open() (err error)
	Add(d Device) (err error)
	AddAll(devices []Device) (err error)
	Remove(ip DeviceIP) (err error)
	Find(ip DeviceIP) (device Device, found bool, err error)
	All() (devices []Device, err error)
	SetActiveUntil(ip DeviceIP, activeUntil time.Time) (err error)

	// Modify the active time by the delta, taking into account the baseTime.
	// Typically the baseTIme is "now".
	// activeTime := max(max(activeTime, baseTime) + delta, baseTime)
	ModifyActiveUntil(ip DeviceIP, delta time.Duration, baseTime time.Time) (err error)
	Close() (err error)
}
