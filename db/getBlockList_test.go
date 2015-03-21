package db

import (
	"testing"
	"time"
)

// GetBlockList test case
type gbltc struct {
	date      string
	probe     string
	active    int
	goodUntil string
	special   string
}

var getBlockListTestCases = []gbltc{
	gbltc{"3/3/15", "1:00PM", 0, "4:00PM", ""},        // Tuesday, schoolday
	gbltc{"3/3/15", "5:00PM", 3, "8:00PM", ""},        // Tuesday, schoolday
	gbltc{"3/3/15", "8:00PM", 1, "10:00PM", "extend"}, // Tuesday, schoolday One exended
	gbltc{"3/3/15", "9:00PM", 0, "4:00PM", ""},        // Tuesday, schoolday
}

func TestGetBlockList(t *testing.T) {
	db := NewSQLDB(":memory:")
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
	calendar, err := NewCalendar(calendarConfig)
	if err != nil {
		t.Errorf("NewCalendar(%v) = %v", calendarConfig, err)
		return
	}
	tc, ok := calendar.(*timeClock)
	if !ok {
		t.Errorf("calender is not a *timeClock. %v", calendar)
		return
	}

	for i, gbltc := range getBlockListTestCases {
		date, err := ParseDate(gbltc.date, tc.location)
		if err != nil {
			t.Errorf("case %d. ParseDate(%s,%s) = %v", i, gbltc.date, tc.location, err)
		}
		probeTimeOfDay, err := ParseTimeOfDay(gbltc.probe)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, gbltc.probe, err)
			continue
		}
		probeTime := tc.mergeDateAndTimeOfDay(date, probeTimeOfDay)

		goodUntilTimeOfDay, err := ParseTimeOfDay(gbltc.goodUntil)
		if err != nil {
			t.Errorf("case %d: ParseTimeOfDay(%v) = %v", i, gbltc.goodUntil, err)
			continue
		}
		goodUntilDate := date
		if goodUntilTimeOfDay.Before(probeTimeOfDay) {
			goodUntilDate = beginningOfNextDay(date)
		}
		expectedGoodUntilTime := tc.mergeDateAndTimeOfDay(goodUntilDate, goodUntilTimeOfDay)

		if gbltc.special == "extend" {
			err = db.SetActiveUntil(ParseDeviceIP("192.168.4.100"), expectedGoodUntilTime)
			if err != nil {
				t.Errorf("case %d: db.SetActiveUntil(%v) = %v", gbltc.date, expectedGoodUntilTime, err)
			}
		}

		blocked, goodUntil, err := GetBlockList(db, calendar, probeTime)
		if err != nil {
			t.Errorf("case %d: %s GetBlockList(%v) = %v", i, gbltc.date, gbltc.probe, err)
			continue
		}
		if !goodUntil.Equal(expectedGoodUntilTime) {
			t.Errorf("case %d: %s GetBlockList(%v) = %v, expected %v", i, gbltc.date, gbltc.probe, goodUntil, expectedGoodUntilTime)
		}

		active := 3 - len(blocked)
		if active != gbltc.active {
			t.Errorf("case %d: %s GetBlockList(%v) -> %v active %v, expected %v", i, gbltc.date, gbltc.probe, blocked, active, gbltc.active)
		}

		// Undo the special modification.
		if gbltc.special == "extend" {
			err = db.SetActiveUntil(ParseDeviceIP("192.168.4.100"), time.Time{})
			if err != nil {
				t.Errorf("case %d: db.SetActiveUntil(%v) = %v", gbltc.date, time.Time{}, err)
			}
		}
	}
}
