package watcher

import (
	"log"
	"net"
	"time"

	"github.com/jackpal/SeattleSnowman/db"
	"github.com/jackpal/SeattleSnowman/router"
)

// Internal implementation of the Watcher.
type firewallUpdater struct {
	db           db.DB
	calendar     db.Calendar
	firewall     router.Firewall
	addressGroup string
	goodUntil    time.Time
}

func (f *firewallUpdater) getBlockList() (blocked []db.DeviceIP, goodUntil time.Time, err error) {
	return db.GetBlockList(f.db, f.calendar, time.Now())
}

func (f *firewallUpdater) updateFirewall() (newWakeTime bool, err error) {
	blocked, goodUntil, err := f.getBlockList()
	log.Printf("updateFirewall(%v) goodUntil %s",
		blocked, goodUntil.Format(time.Kitchen))
	if err != nil {
		return
	}
	newWakeTime = !f.goodUntil.Equal(goodUntil)
	f.goodUntil = goodUntil
	var ips router.IPs
	for _, deviceIP := range blocked {
		ips = append(ips, net.IP(deviceIP))
	}
	log.Printf("new blocklist: %v", ips)
	err = f.firewall.SetAddressGroup(f.addressGroup, ips)
	return
}

type Watcher struct {
	db       db.DB
	wi       *firewallUpdater
	commands chan func(*firewallUpdater)
	timeout  chan bool
}

func NewWatcher(db db.DB, calendar db.Calendar, firewall router.Firewall,
	addressGroup string) (w *Watcher) {
	return &Watcher{
		db,
		&firewallUpdater{db, calendar, firewall, addressGroup, time.Time{}},
		make(chan func(*firewallUpdater), 1),
		make(chan bool, 1),
	}
}

func (w *Watcher) pingFirewall() {
	w.commands <- func(wi *firewallUpdater) {
		oldWakeTime := wi.goodUntil
		_, err := wi.updateFirewall()
		if err != nil {
			log.Printf("Error updating firewall: %v", err)
		}
		w.maybeScheduleTimeout(oldWakeTime, w.wi.goodUntil)
	}
}

func (w *Watcher) pingIfNoError(errIn error) (err error) {
	err = errIn
	if err == nil {
		w.pingFirewall()
	}
	return
}

func (w *Watcher) AddDevices(devices []db.Device) (err error) {
	err = w.pingIfNoError(w.db.AddAll(devices))
	return
}

func (w *Watcher) AddDevice(device db.Device) (err error) {
	err = w.pingIfNoError(w.db.Add(device))
	return
}

func (w *Watcher) ModifyActiveUntil(ip db.DeviceIP, delta time.Duration) (err error) {
	err = w.pingIfNoError(w.db.ModifyActiveUntil(ip, delta, time.Now()))
	return
}

func (w *Watcher) SetActiveUntil(ip db.DeviceIP, activeUntil time.Time) (err error) {
	err = w.pingIfNoError(w.db.SetActiveUntil(ip, activeUntil))
	return
}

func (w *Watcher) maybeScheduleTimeout(oldWakeTime time.Time, wakeTime time.Time) {
	if !wakeTime.IsZero() &&
		(oldWakeTime.IsZero() || wakeTime.Before(oldWakeTime)) {
		sleepTime := wakeTime.Sub(time.Now())
		go func() {
			time.Sleep(sleepTime)
			w.timeout <- true
		}()
	}
}

func (w *Watcher) Close() (err error) {
	close(w.commands)
	return
}

func (w *Watcher) State() (state []db.Device, err error) {
	state, err = w.db.All()
	if err != nil {
		return
	}
	// To simplify UI, treat all times before now as zero.
	state = zeroTimesBefore(state, time.Now())
	return
}

func zeroTimesBefore(state []db.Device, t time.Time) (stateZ []db.Device) {
	for _, d := range state {
		activeUntil := d.ActiveUntil
		if !activeUntil.IsZero() && activeUntil.Before(t) {
			d.ActiveUntil = time.Time{}
		}
		stateZ = append(stateZ, d)
	}
	return
}

func (w *Watcher) BlockList() (blockList []db.DeviceIP, goodUntil time.Time, err error) {
	return w.wi.getBlockList()
}

func (w *Watcher) Start() (err error) {
	go w.loop()
	return
}

func (w *Watcher) loop() {
	w.updateFirewall()
	for {
		select {
		case command := <-w.commands:
			command(w.wi)
		case <-w.timeout:
			w.updateFirewall()
		}
	}
}

func (w *Watcher) updateFirewall() {
	oldWakeTime := w.wi.goodUntil
	w.wi.updateFirewall()
	w.maybeScheduleTimeout(oldWakeTime, w.wi.goodUntil)
}
