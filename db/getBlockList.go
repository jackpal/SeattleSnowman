// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package db

import "time"

func maxTime(a time.Time, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.Before(b) {
		return b
	}
	return a
}

func minTime(a time.Time, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if a.Before(b) {
		return a
	}
	return b
}

func GetBlockList(db DB, calendar Calendar, atTime time.Time) (blocked []DeviceIP, goodUntil time.Time, err error) {
	all, err := db.All()
	if err != nil {
		return
	}
	calendarActive, calendarPeriod := calendar.RuleAt(atTime)
	for _, d := range all {
		deviceActiveEnd := time.Time{}
		if calendarActive {
			deviceActiveEnd = calendarPeriod.End
		}
		dbActiveUntil := d.ActiveUntil
		if !dbActiveUntil.IsZero() && dbActiveUntil.After(atTime) {
			deviceActiveEnd = maxTime(deviceActiveEnd, dbActiveUntil)
		}
		if atTime.Before(deviceActiveEnd) {
			goodUntil = minTime(goodUntil, deviceActiveEnd)
		} else {
			blocked = append(blocked, d.IP)
		}
	}
	if goodUntil.IsZero() && !calendarActive {
		goodUntil = calendarPeriod.End
	}
	return
}
