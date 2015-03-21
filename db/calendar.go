package db

import "time"

type TimePeriod struct {
	Start time.Time
	End   time.Time
}

func (tp TimePeriod) Includes(u time.Time) bool {
	return u.Sub(tp.Start) >= 0 && tp.End.Sub(u) > 0
}

func (tp TimePeriod) Equal(o TimePeriod) bool {
	return tp.Start.Equal(o.Start) && tp.End.Equal(o.End)
}

type TimeOfDayPeriod TimePeriod

func (tp TimeOfDayPeriod) Equal(o TimeOfDayPeriod) bool {
	return TimePeriod(tp).Equal(TimePeriod(o))
}

func (d TimeOfDayPeriod) String() string {
	return d.Start.Format(time.Kitchen) + " - " + d.End.Format(time.Kitchen)
}

type DatePeriod TimePeriod

const DateFormat = "1/2/06"

func (tp DatePeriod) Equal(o DatePeriod) bool {
	return TimePeriod(tp).Equal(TimePeriod(o))
}

func (d DatePeriod) String() string {
	return d.Start.Format(DateFormat) + " - " + d.End.Format(DateFormat)
}

type Calendar interface {
	RuleAt(t time.Time) (isOn bool, period TimePeriod)
}

type timeClock struct {
	location       *time.Location
	schoolDayHours TimeOfDayPeriod // Monday hours
	vacationHours  TimeOfDayPeriod // Saturday hours
	holidays       []DatePeriod
}

// This configuration information should be in a database someday, but for now
// it's convenient to set it at compile time.

type TimeOfDayPeriodConfig struct {
	StartTime string
	EndTime   string
}

type DateRangeConfig struct {
	StartDay string
	// This is inclusive
	EndDay string
}

type CalendarConfig struct {
	Location       string
	SchoolDayHours TimeOfDayPeriodConfig
	VacationHours  TimeOfDayPeriodConfig
	Holidays       []DateRangeConfig
}

func NewCalendar(cc *CalendarConfig) (tc Calendar, err error) {
	tc, err = ParseTimeClock(cc)
	return
}

func ParseTimeClock(cc *CalendarConfig) (tc *timeClock, err error) {
	location, err := time.LoadLocation(cc.Location)
	if err != nil {
		return
	}
	schoolDayHours, err := ParseTimeOfDayPeriod(cc.SchoolDayHours)
	if err != nil {
		return
	}
	vacationHours, err := ParseTimeOfDayPeriod(cc.VacationHours)
	if err != nil {
		return
	}
	holidays, err := parseHolidays(cc.Holidays, location)
	if err != nil {
		return
	}
	tc = &timeClock{location, schoolDayHours, vacationHours, holidays}
	return
}

func ParseTimeOfDay(tod string) (t time.Time, err error) {
	return time.Parse(time.Kitchen, tod)
}

func ParseTimeOfDayPeriod(todc TimeOfDayPeriodConfig) (tp TimeOfDayPeriod, err error) {
	tp.Start, err = ParseTimeOfDay(todc.StartTime)
	if err != nil {
		return
	}
	tp.End, err = ParseTimeOfDay(todc.EndTime)
	if err != nil {
		return
	}
	return
}

func beginningOfPreviousDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day-1, 0, 0, 0, 0, t.Location())
}

func beginningOfNextDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day+1, 0, 0, 0, 0, t.Location())
}

func parseHolidays(hc []DateRangeConfig, location *time.Location) (holidays []DatePeriod, err error) {
	for _, h := range hc {
		var holiday DatePeriod
		holiday, err = ParseDateRange(h, location)
		if err != nil {
			return
		}
		holidays = append(holidays, holiday)
	}
	return
}

func ParseDate(date string, location *time.Location) (t time.Time, err error) {
	return time.ParseInLocation(DateFormat, date, location)
}

func ParseDateRange(dr DateRangeConfig, location *time.Location) (dp DatePeriod, err error) {
	var start, end time.Time
	start, err = ParseDate(dr.StartDay, location)
	if err != nil {
		return
	}
	end, err = ParseDate(dr.EndDay, location)
	if err != nil {
		return
	}
	dp = DatePeriod{start, beginningOfNextDay(end)}
	return
}

func (tc *timeClock) RuleAt(t time.Time) (isOn bool, period TimePeriod) {
	activeTime := tc.activeTimeFor(t)
	if activeTime.Includes(t) {
		isOn, period = true, activeTime
		return
	}
	if t.Before(activeTime.Start) {
		prevDayEnd := tc.endTimeFor(beginningOfPreviousDay(t))
		return false, TimePeriod{prevDayEnd, activeTime.Start}
	} else {
		nextDayStart := tc.startTimeFor(beginningOfNextDay(t))
		return false, TimePeriod{activeTime.End, nextDayStart}
	}
}

func (tc *timeClock) activeTimeFor(t time.Time) TimePeriod {
	return TimePeriod{tc.startTimeFor(t), tc.endTimeFor(t)}
}

func (tc *timeClock) isSchoolDay(t time.Time) bool {
	if tc.isHoliday(t) {
		return false
	}
	weekDay := t.Weekday()
	return weekDay >= time.Monday && weekDay <= time.Friday
}

func (tc *timeClock) isSchoolNight(t time.Time) bool {
	return tc.isSchoolDay(beginningOfNextDay(t))
}

func (tc *timeClock) startTimeOfDayFor(t time.Time) time.Time {
	return tc.activeHoursForDayType(tc.isSchoolDay(t)).Start
}

func (tc *timeClock) endTimeOfDayFor(t time.Time) time.Time {
	return tc.activeHoursForDayType(tc.isSchoolNight(t)).End
}

func (tc *timeClock) startTimeFor(t time.Time) time.Time {
	return tc.mergeDateAndTimeOfDay(t, tc.startTimeOfDayFor(t))
}

func (tc *timeClock) endTimeFor(t time.Time) time.Time {
	return tc.mergeDateAndTimeOfDay(t, tc.endTimeOfDayFor(t))
}

func (tc *timeClock) isHoliday(t time.Time) bool {
	for _, h := range tc.holidays {
		includes := TimePeriod(h).Includes(t)
		if includes {
			return true
		}
	}
	return false
}

func (tc *timeClock) activeHoursForDayType(isSchoolDay bool) (tod TimeOfDayPeriod) {
	if isSchoolDay {
		return tc.schoolDayHours
	} else {
		return tc.vacationHours
	}
	return
}

func (tc *timeClock) mergeDateAndTimeOfDay(date time.Time, timeOfDay time.Time) time.Time {
	year, month, day := date.Date()
	hour, minute, second := timeOfDay.Hour(), timeOfDay.Minute(), timeOfDay.Second()
	return time.Date(year, month, day, hour, minute, second, 0, tc.location)
}
