// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package db

import (
	"testing"
	"time"
)

// TimePeriod Includes test case struct
type tpitc struct {
	start    string
	end      string
	test     string
	expected bool
}

var timePeriodTest = []tpitc{
	tpitc{"2:00PM", "3:00PM", "1:00PM", false},
	tpitc{"2:00PM", "3:00PM", "2:00PM", true},
	tpitc{"2:00PM", "3:00PM", "2:30PM", true},
	tpitc{"2:00PM", "3:00PM", "3:00PM", false},
	tpitc{"2:00PM", "3:00PM", "4:00PM", false},
}

func TestTimePeriod(t *testing.T) {
	for i, tc := range timePeriodTest {
		todc := TimeOfDayPeriodConfig{tc.start, tc.end}
		todp, err := ParseTimeOfDayPeriod(todc)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDayPeriod(%v) = %v", i, todc, err)
			continue
		}
		testTime, err := ParseTimeOfDay(tc.test)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, tc.test, err)
			continue
		}
		result := TimePeriod(todp).Includes(testTime)
		if result != tc.expected {
			t.Errorf("case %d: %v.Includes(%v) = %v, expected %v", i, todp, tc.test, result, tc.expected)
		}
	}
}

var datePeriodTestLocation = time.UTC

// DatePeriod test case
var datePeriodTest = []tpitc{
	tpitc{"12/23/11", "01/02/12", "12/22/11", false},
	tpitc{"12/23/11", "01/02/12", "12/23/11", true},
	tpitc{"12/23/11", "01/02/12", "12/25/11", true},
	tpitc{"12/23/11", "01/02/12", "01/02/12", true}, // Last Day included
	tpitc{"12/23/11", "01/02/12", "01/03/12", false},
}

func TestDatePeriod(t *testing.T) {
	for i, tc := range datePeriodTest {
		drc := DateRangeConfig{tc.start, tc.end}
		dp, err := ParseDateRange(drc, datePeriodTestLocation)
		if err != nil {
			t.Errorf("case %d: ParseDateRange(%v) = %v", i, drc, err)
			continue
		}
		testTime, err := ParseDate(tc.test, datePeriodTestLocation)
		if err != nil {
			t.Errorf("case %d: ParseDate(%v, %v) = %v", i, tc.test, datePeriodTestLocation, err)
			continue
		}
		result := TimePeriod(dp).Includes(testTime)
		if result != tc.expected {
			t.Errorf("case %d: %v.Includes(%v) = %v, expected %v", i, dp, tc.test, result, tc.expected)
		}
	}
}

var calendarConfig = &CalendarConfig{
	"America/Los_Angeles",
	TimeOfDayPeriodConfig{"4:00PM", "8:00PM"}, // SchoolDay
	TimeOfDayPeriodConfig{"1:00PM", "9:00PM"}, // VacationDay
	[]DateRangeConfig{
		DateRangeConfig{"4/6/15", "4/10/15"},
		DateRangeConfig{"5/22/15", "5/25/15"},
	},
}

func TestCalendar(t *testing.T) {
	calendar, err := NewCalendar(calendarConfig)
	if err != nil {
		t.Errorf("NewCalendar(%v) = %v", calendarConfig, err)
		return
	}
	if tc, ok := calendar.(*timeClock); ok {
		testTimeClock(t, tc)
	} else {
		t.Errorf("calender is not a *timeClock. %v", calendar)
	}
}

// school day test case
type sdtc struct {
	date          string
	isHoliday     bool
	isSchoolDay   bool
	isSchoolNight bool
}

var schoolDayTestCases = []sdtc{
	sdtc{"3/1/15", false, false, true},  // Sun
	sdtc{"3/2/15", false, true, true},   // Mon
	sdtc{"3/3/15", false, true, true},   // Tue
	sdtc{"3/6/15", false, true, false},  // Fri
	sdtc{"3/7/15", false, false, false}, // Sat
	sdtc{"4/1/15", false, true, true},   // Wed
	sdtc{"4/3/15", false, true, false},  // Fri
	sdtc{"4/4/15", false, false, false}, // Sat
	sdtc{"4/5/15", false, false, false}, // Sun before holiday
	sdtc{"4/6/15", true, false, false},  // Mon and holiday
	sdtc{"4/7/15", true, false, false},  // Tue and holiday
	sdtc{"4/10/15", true, false, false}, // Fri and holiday
	sdtc{"4/13/15", false, true, true},  // Mon
	sdtc{"5/10/15", false, false, true}, // Sun and not holiday
	sdtc{"5/24/15", true, false, false}, // Sun and holiday
}

func testSchoolDays(t *testing.T, tc *timeClock) {
	for i, sdtc := range schoolDayTestCases {
		date, err := ParseDate(sdtc.date, tc.location)
		if err != nil {
			t.Errorf("case %d: ParseDate(%v, %v) = %v", i, sdtc.date, tc.location, err)
			continue
		}
		{
			isHoliday := tc.isHoliday(date)
			if isHoliday != sdtc.isHoliday {
				t.Errorf("case %d: tc.isHoliday(%v) = %v, expected %v", i,
					sdtc.date, isHoliday, sdtc.isHoliday)
			}
		}
		{
			isSchoolDay := tc.isSchoolDay(date)
			if isSchoolDay != sdtc.isSchoolDay {
				t.Errorf("case %d: tc.isSchoolDay(%v) = %v, expected %v", i,
					sdtc.date, isSchoolDay, sdtc.isSchoolDay)
			}
		}
		{
			isSchoolNight := tc.isSchoolNight(date)
			if isSchoolNight != sdtc.isSchoolNight {
				t.Errorf("case %d: tc.isSchoolNight(%v) = %v, expected %v", i,
					sdtc.date, isSchoolNight, sdtc.isSchoolNight)
			}
		}
	}
}

type attc struct {
	date  string
	probe string
	isOn  bool
	begin string
	end   string
}

var activeTimeTestCases = []attc{
	// Sunday vacation Day, School Night
	attc{"03/01/15", "12:00PM", false, "9:00PM", "1:00PM"},
	attc{"03/01/15", "1:00PM", true, "1:00PM", "8:00PM"},
	attc{"03/01/15", "8:00PM", false, "8:00PM", "4:00PM"},
	// Monday
	attc{"03/02/15", "3:00PM", false, "8:00PM", "4:00PM"},
	attc{"03/02/15", "4:00PM", true, "4:00PM", "8:00PM"},
	attc{"03/02/15", "8:00PM", false, "8:00PM", "4:00PM"},
	// Tues
	attc{"03/03/15", "3:00PM", false, "8:00PM", "4:00PM"},
	attc{"03/03/15", "4:00PM", true, "4:00PM", "8:00PM"},
	attc{"03/03/15", "8:00PM", false, "8:00PM", "4:00PM"},
	// Fri
	attc{"03/06/15", "3:00PM", false, "8:00PM", "4:00PM"},
	attc{"03/06/15", "4:00PM", true, "4:00PM", "9:00PM"},
	attc{"03/06/15", "9:00PM", false, "9:00PM", "1:00PM"},
	// Sat
	attc{"03/07/15", "12:00PM", false, "9:00PM", "1:00PM"},
	attc{"03/07/15", "1:00PM", true, "1:00PM", "9:00PM"},
	attc{"03/07/15", "9:00PM", false, "9:00PM", "1:00PM"},
}

func testActiveTime(t *testing.T, tc *timeClock) {
	for i, attc := range activeTimeTestCases {
		date, err := ParseDate(attc.date, tc.location)
		if err != nil {
			t.Errorf("case %d. ParseDate(%s,%s) = %v", i, attc.date, tc.location, err)
		}
		probeTimeOfDay, err := ParseTimeOfDay(attc.probe)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, attc.probe, err)
			continue
		}
		probeTime := tc.mergeDateAndTimeOfDay(date, probeTimeOfDay)

		beginTimeOfDay, err := ParseTimeOfDay(attc.begin)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, attc.begin, err)
			continue
		}
		beginDate := date
		if beginTimeOfDay.After(probeTimeOfDay) {
			beginDate = beginningOfPreviousDay(date)
		}
		expectedBeginTime := tc.mergeDateAndTimeOfDay(beginDate, beginTimeOfDay)

		endTimeOfDay, err := ParseTimeOfDay(attc.end)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, attc.end, err)
			continue
		}
		endDate := date
		if endTimeOfDay.Before(probeTimeOfDay) {
			endDate = beginningOfNextDay(date)
		}
		expectedEndTime := tc.mergeDateAndTimeOfDay(endDate, endTimeOfDay)

		expectedPeriod := TimePeriod{expectedBeginTime, expectedEndTime}

		isOn, period := tc.RuleAt(probeTime)
		if isOn != attc.isOn || !period.Equal(expectedPeriod) {
			t.Errorf("case %d: %v tc.RuleAt(%v) = %v,%v, expected %v,%v",
				i, attc.date, attc.probe, isOn, period, attc.isOn, expectedPeriod)
		}
	}
}

func testTimeClock(t *testing.T, tc *timeClock) {
	testActiveHoursForDayType(t, tc)
	testSchoolDays(t, tc)
	testActiveTime(t, tc)
}

func testActiveHoursForDayType(t *testing.T, tc *timeClock) {
	checkEqualTimeOfDayPeriod(t, "tc.activeHoursForDayType(true)", tc.activeHoursForDayType(true), tc.schoolDayHours)
	checkEqualTimeOfDayPeriod(t, "tc.activeHoursForDayType(false)", tc.activeHoursForDayType(false), tc.vacationHours)
}

func checkEqualTimeOfDayPeriod(t *testing.T, label string, a TimeOfDayPeriod, b TimeOfDayPeriod) {
	if !a.Equal(b) {
		t.Errorf("%s = %v (expected %v)", label, a, b)
	}
}

func checkEqualTimePeriod(t *testing.T, label string, a TimePeriod, b TimePeriod) {
	if !a.Equal(b) {
		t.Errorf("%s = %v (expected %v)", label, a, b)
	}
}

func checkEqualDatePeriod(t *testing.T, label string, a DatePeriod, b DatePeriod) {
	if !a.Equal(b) {
		t.Errorf("%s = %v (expected %v)", label, a, b)
	}
}
