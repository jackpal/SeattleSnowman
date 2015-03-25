package db

import (
	"sync"
	"time"
)

// A RAM-based DB does not persist to disk.
func NewRAMDB() (d DB) {
	return &ram{}
}

type ram struct {
	mutex   sync.RWMutex
	devices []Device
}

func (r *ram) Open() (err error) {
	return
}

func (r *ram) Add(d Device) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.devices = append(r.devices, d)
	return
}

func (r *ram) AddAll(devices []Device) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.devices = append(r.devices, devices...)
	return
}

func (r *ram) Remove(ip DeviceIP) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	i := r.find(ip)
	if i >= 0 {
		r.devices = append(r.devices[:i], r.devices[i+1:]...)
	}
	return
}

func (r *ram) Find(ip DeviceIP) (device Device, found bool, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	index := r.find(ip)
	found = index >= 0
	if found {
		device = r.devices[index]
	}
	return
}

func (r *ram) find(ip DeviceIP) (i int) {
	for i, d := range r.devices {
		if d.IP.Equal(ip) {
			return i
		}
	}
	return -1
}

func (r *ram) All() (devices []Device, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	devices = r.devices[:]
	return
}

func (r *ram) SetActiveUntil(ip DeviceIP, activeUntil time.Time) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	i := r.find(ip)
	if i >= 0 {
		r.devices[i].ActiveUntil = activeUntil
	}
	return
}

func (r *ram) ModifyActiveUntil(ip DeviceIP, delta time.Duration,
	baseTime time.Time) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	i := r.find(ip)
	if i >= 0 {
		d := &r.devices[i]
		d.ActiveUntil = modifyActiveUntil(d.ActiveUntil, delta, baseTime)
	}
	return
}

func (r *ram) Close() (err error) {
	return
}
