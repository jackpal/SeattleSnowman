// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/jackpal/SeattleSnowman/db"
	"github.com/jackpal/SeattleSnowman/router"
	"github.com/jackpal/SeattleSnowman/watcher"
)

type Configuration struct {
	Port                 int    // Port to serve from.
	AddressGroup         string // Router Filter address group.
	Database             string // Path to database file.
	RouterAddress        string // Router ssh address (name:port, port is optional);
	RouterPrivateKeyPath string // Router ssh private key file.
	Calendar             db.CalendarConfig
	Devices              []db.Device
}

var configFile = flag.String("config", "config.json", "Configuration file.")

var watch *watcher.Watcher

func newWatcher(config *Configuration) (w *watcher.Watcher, err error) {
	database := db.NewSQLDB(config.Database)
	err = database.Open()
	if err != nil {
		return
	}
	err = maybeAddDevices(database, config.Devices)
	if err != nil {
		return
	}
	calendar, err := db.NewCalendar(&config.Calendar)
	if err != nil {
		return
	}
	firewall := router.NewEdgeRouterFirewall(config.RouterAddress, config.RouterPrivateKeyPath)
	w = watcher.NewWatcher(database, calendar, firewall, config.AddressGroup)
	return
}

func maybeAddDevices(db db.DB, devices []db.Device) (err error) {
	for _, device := range devices {
		var found bool
		_, found, err = db.Find(device.IP)
		if err != nil {
			return
		}
		if !found {
			err = db.Add(device)
		}
		if err != nil {
			return
		}
	}
	return
}

func writeJSON(w http.ResponseWriter, jsonData interface{}, err error) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(jsonData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func handleDeviceListImp(r *http.Request) (state []db.Device, err error) {
	if r.Method != "GET" {
		err = fmt.Errorf("Method != GET")
		return
	}
	state, err = watch.State()
	return
}

func handleDeviceList(w http.ResponseWriter, r *http.Request) {
	state, err := handleDeviceListImp(r)
	writeJSON(w, state, err)
}

func parseIPs(r *http.Request) (ips router.IPs, err error) {
	ipValues := r.URL.Query()["ip"]
	for _, ipValue := range ipValues {
		ip := net.ParseIP(ipValue)
		if ip == nil {
			err = fmt.Errorf("Could not parse IP value %q", ipValue)
			return
		}
		ips = append(ips, ip)
	}
	return
}

func handleBlockListImp(r *http.Request) (blocked []db.DeviceIP, goodUntil time.Time, err error) {
	if r.Method != "GET" {
		err = fmt.Errorf("Method != GET")
		return
	}
	blocked, goodUntil, err = watch.BlockList()
	return
}

func handleBlockList(w http.ResponseWriter, r *http.Request) {
	blocked, goodUntil, err := handleBlockListImp(r)
	var a = make(map[string]interface{})
	a["blocked"] = blocked
	a["goodUntil"] = goodUntil
	writeJSON(w, a, err)
}

func handleAddDevice(w http.ResponseWriter, r *http.Request) {
	err := addDeviceImp(r)
	writeJSON(w, nil, err)
}

func addDeviceImp(r *http.Request) (err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Method != POST")
		return
	}
	ip := r.FormValue("ip")
	name := r.FormValue("name")
	err = watch.AddDevice(db.NewDevice(ip, name))
	return
}

func handleBlock(w http.ResponseWriter, r *http.Request) {
	err := setActiveUntilImp(r, time.Time{})
	writeJSON(w, nil, err)
}

func handleUnblock(w http.ResponseWriter, r *http.Request) {
	err := handleUnblockImp(r)
	writeJSON(w, nil, err)
}

func handleUnblockImp(r *http.Request) (err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Method != POST")
		return
	}
	hours := r.FormValue("hours")
	hoursInt, err := strconv.ParseInt(hours, 10, 0)
	if err != nil {
		return
	}
	blockTime := time.Now().Add(time.Duration(hoursInt) * time.Hour)
	err = setActiveUntilImp(r, blockTime)
	return
}

func modifyActiveUntilImp(r *http.Request) (err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Must use POST")
		return
	}
	ip := r.FormValue("ip")
	if ip == "" {
		err = fmt.Errorf("Missing ip parameter")
		return
	}
	deviceIP := db.ParseDeviceIP(ip)
	delta, err := time.ParseDuration(r.FormValue("delta"))
	if err != nil {
		return
	}
	err = watch.ModifyActiveUntil(deviceIP, delta)
	if err != nil {
		return
	}
	return
}

func handleModifyActiveUntil(w http.ResponseWriter, r *http.Request) {
	err := modifyActiveUntilImp(r)
	writeJSON(w, nil, err)
}

func setActiveUntilImp(r *http.Request, activeUntil time.Time) (err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Must use POST")
		return
	}
	ip := r.FormValue("ip")
	deviceIP := db.ParseDeviceIP(ip)
	err = watch.SetActiveUntil(deviceIP, activeUntil)
	if err != nil {
		return
	}
	return
}

func handleUploadDevices(w http.ResponseWriter, r *http.Request) {
	err := uploadDevicesImp(r)
	writeJSON(w, nil, err)
}

func uploadDevicesImp(r *http.Request) (err error) {
	if r.Method != "POST" {
		err = fmt.Errorf("Method != POST")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return
	}
	defer file.Close()
	devices, err := readDeviceConfig(file)
	if err != nil {
		return
	}
	dbDevices, err := deviceConfigToDevices(devices)
	if err != nil {
		return
	}
	err = watch.AddDevices(dbDevices)
	return
}

func deviceConfigToDevices(dces []*DeviceConfigEntry) (devices []db.Device, err error) {
	for _, dce := range dces {
		if dce.Kids == "" {
			continue
		}
		d := db.NewDevice(dce.IP, dce.Name)
		devices = append(devices, d)
	}
	return
}

func handleDevices(w http.ResponseWriter, r *http.Request) {
	err := handleDevicesImp(w, r)
	if err != nil {
		log.Printf("handleMainPageImp() = %v", err)
	}
}

func timeIsZero(t time.Time) bool {
	return t.IsZero()
}

func kitchen(t time.Time) string {
	return t.Format(time.Kitchen)
}

// Sort by device name.
type byName []db.Device

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }

func handleDevicesImp(w http.ResponseWriter, r *http.Request) (err error) {
	funcMap := template.FuncMap{
		"timeIsZero": timeIsZero,
		"kitchen":    kitchen,
	}
	tmpl, err := template.New("devices.html").Funcs(funcMap).ParseFiles("templates/devices.html")
	if err != nil {
		return
	}

	devices, err := watch.State()
	if err != nil {
		return
	}
	sort.Sort(byName(devices))
	err = tmpl.Execute(w, devices)
	return
}

func main() {
	log.Printf("Seattle Snowman starting")
	defer log.Printf("Seattle Snowman ending")
	err := mainLoop()
	if err != nil {
		log.Printf("mainLoop() error %v", err)
	}
}

func loadConfig() (config *Configuration, err error) {
	file, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return
	}
	var c Configuration
	err = json.Unmarshal(file, &c)
	if err != nil {
		return
	}
	log.Printf("Configuration: %+v", c)
	config = &c
	return
}

func mainLoop() (err error) {
	flag.Parse()
	config, err := loadConfig()
	if err != nil {
		return
	}
	watch, err = newWatcher(config)
	if err != nil {
		log.Printf("newWatcher() = %v", err)
		return
	}
	defer watch.Close()
	err = watch.Start()
	if err != nil {
		return
	}

	http.HandleFunc("/addDevice", handleAddDevice)
	http.HandleFunc("/blockList", handleBlockList)
	http.HandleFunc("/deviceList", handleDeviceList)
	http.HandleFunc("/block", handleBlock)
	http.HandleFunc("/unblock", handleUnblock)
	http.HandleFunc("/modifyActiveUntil", handleModifyActiveUntil)
	http.HandleFunc("/uploadDevices", handleUploadDevices)
	http.HandleFunc("/devices.html", handleDevices)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	address := net.JoinHostPort("", strconv.Itoa(config.Port))
	err = http.ListenAndServe(address, nil)
	return
}
