// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package db

import (
	"fmt"
	"testing"
	"time"
)

func TestSQLDB(t *testing.T) {
	db := NewSQLDB(":memory:")
	exerciseDB(t, db)
}

func exerciseDB(t *testing.T, db DB) {
	err := db.Open()
	if err != nil {
		t.Errorf("db.Open() = %v", err)
		return
	}
	defer testClose(t, db)
	err = loadDB(t, db)
	if err != nil {
		return
	}
	devices, err := db.All()
	if err != nil || len(devices) != 3 {
		t.Errorf("db.All() = %v", err)
		return
	}
	martianIP := ParseDeviceIP("10.10.10.10")
	err = db.Add(Device{martianIP, "martian", time.Time{}})
	if err != nil {
		t.Errorf("db.Add(\"martian\") = %v", err)
		return
	}
	testTime := time.Unix(1234567, 0)
	err = db.SetActiveUntil(martianIP, testTime)
	if err != nil {
		t.Errorf("db.SetActiveUntil(%v,%v) = %v", martianIP, testTime, err)
		return
	}

	activeUntil, err := getActiveUntilHelper(db, martianIP)
	if err != nil {
		t.Errorf("getActiveUntilHelper(%v) = %v", martianIP, err)
		return
	}

	if !activeUntil.Equal(testTime) {
		t.Errorf("db.SetActiveUntil(%v) = %v", testTime, activeUntil)
	}

	err = testModifyActive(t, db, martianIP)
	if err != nil {
		return
	}

	err = db.Remove(martianIP)
	if err != nil {
		t.Errorf("db.Remove(\"martian\") = %v", err)
		return
	}

	device, found, err := db.Find(martianIP)
	if err != nil || found {
		t.Errorf("after remove db.Find(\"martian\") = %v, %v, %v", device, found, err)
		return
	}
}

func getActiveUntilHelper(db DB, ip DeviceIP) (activeUntil time.Time, err error) {
	device, found, err := db.Find(ip)
	if err != nil || !found || !device.IP.Equal(ip) {
		err = fmt.Errorf("db.Find(\"%v\") = %v, %v, %v", ip, device, found, err)
		return
	}
	activeUntil = device.ActiveUntil
	return
}

type modifyActiveTestCase struct {
	oldActive      string // Time in Kitchen format
	delta          string // Duration
	base           string // Time in Kitchen format
	expectedActive string // Time in Kitchen format
}

var modifyActiveTestCases = []modifyActiveTestCase{
	modifyActiveTestCase{"4:00PM", "1h", "1:00PM", "5:00PM"},
	modifyActiveTestCase{"4:00PM", "-1h", "1:00PM", "3:00PM"},
	modifyActiveTestCase{"1:00PM", "1h", "3:00PM", "4:00PM"},
	modifyActiveTestCase{"1:00PM", "-1h", "1:00PM", "1:00PM"},
}

func testModifyActive(t *testing.T, db DB, ip DeviceIP) (err error) {
	for i, testCase := range modifyActiveTestCases {
		err = testModifyActiveTestCase(t, db, ip, testCase)
		if err != nil {
			t.Errorf("modifyActiveTestCases[%d] == %v -> error: %v", i, testCase, err)
			return
		}
	}
	return
}

func parseTimeHelper(k string, v string) (t time.Time, err error) {
	t, err = time.Parse(time.Kitchen, v)
	if err != nil {
		err = fmt.Errorf("%v : time.Parse(%v, kitchen) = %v", k, v, err)
	}
	return
}

func parseDurationHelper(k string, v string) (d time.Duration, err error) {
	d, err = time.ParseDuration(v)
	if err != nil {
		err = fmt.Errorf("%v : time.ParseDuration(%v) = %v", k, v, err)
	}
	return
}

func testModifyActiveTestCase(t *testing.T, db DB, ip DeviceIP, tc modifyActiveTestCase) (err error) {
	oldActive, err := parseTimeHelper("oldActive", tc.oldActive)
	if err != nil {
		return
	}
	delta, err := parseDurationHelper("delta", tc.delta)
	if err != nil {
		return
	}
	base, err := parseTimeHelper("base", tc.base)
	if err != nil {
		return
	}
	expectedActive, err := parseTimeHelper("expectedActive", tc.expectedActive)
	if err != nil {
		return
	}
	err = db.SetActiveUntil(ip, oldActive)
	if err != nil {
		err = fmt.Errorf("db.SetActiveUntil(%v,%v) = %v", ip, oldActive, err)
		return
	}
	err = db.ModifyActiveUntil(ip, delta, base)
	if err != nil {
		t.Errorf("db.ModifyActiveUntil(%v,%v,%v) = %v", ip, delta, base, err)
		return
	}
	newActiveUntil, err := getActiveUntilHelper(db, ip)
	if err != nil {
		return
	}
	if !newActiveUntil.Equal(expectedActive) {
		t.Errorf("db.ModifyActiveUntil(%v,%v,%v) = %v", ip, delta, base, newActiveUntil)
		err = fmt.Errorf("newActiveUntil = %v, expectedActive = %v", newActiveUntil, expectedActive)
		return
	}
	return
}

func testClose(t *testing.T, db DB) {
	err := db.Close()
	if err != nil {
		t.Errorf("db.Close() = %v", err)
	}
}

type entry struct {
	ip   string
	name string
}

var testData = []entry{
	{"192.168.4.100", "Able"},
	{"192.168.4.101", "Baker"},
	{"192.168.4.102", "Charlie"},
}

func loadDB(t *testing.T, db DB) (err error) {
	return loadDBData(t, db, testData)
}

func convertEntriesToDevices(entries []entry) (devices []Device) {
	for _, e := range entries {
		d := Device{ParseDeviceIP(e.ip), e.name, time.Time{}}
		devices = append(devices, d)
	}
	return
}

func loadDBData(t *testing.T, db DB, entries []entry) (err error) {
	devices := convertEntriesToDevices(entries)
	err = db.AddAll(devices)
	return
}
