package trading

import "time"

// B3CalendarVersion is intentionally finite. Unknown years fail closed until a
// reviewed exchange calendar is added. Sources are recorded in docs/web-research.nd.
const B3CalendarVersion = "B3 equities 2026 / OC005 + OC055 / reviewed 2026-11-03"

var b3Holidays = map[string]bool{
	"2026-01-01": true,
	"2026-02-16": true,
	"2026-02-17": true,
	"2026-04-03": true,
	"2026-04-21": true,
	"2026-05-01": true,
	"2026-06-04": true,
	"2026-09-07": true,
	"2026-10-12": true,
	"2026-11-02": true,
	"2026-11-20": true,
	"2026-12-24": true,
	"2026-12-25": true,
	"2026-12-31": true,
}

func SessionBounds(market string, at time.Time) (time.Time, time.Time, bool) {
	zone := "America/New_York"
	hour, minute, closeHour, closeMinute := 9, 30, 16, 0
	if market == "B3" {
		zone = "America/Sao_Paulo"
		hour, minute, closeHour, closeMinute = 10, 0, 16, 55
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	t := at.In(loc)
	date := t.Format("2006-01-02")
	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return time.Time{}, time.Time{}, false
	}
	if market == "B3" {
		if t.Year() != 2026 || b3Holidays[date] {
			return time.Time{}, time.Time{}, false
		}
		if date == "2026-02-18" {
			hour = 13
		}
		if date < "2026-03-09" || date >= "2026-11-03" {
			closeHour = 17
		}
	}
	start := time.Date(t.Year(), t.Month(), t.Day(), hour, minute, 0, 0, loc)
	end := time.Date(t.Year(), t.Month(), t.Day(), closeHour, closeMinute, 0, 0, loc)
	return start, end, true
}

func Session(market string, t time.Time) bool {
	start, end, ok := SessionBounds(market, t)
	return ok && !t.Before(start) && t.Before(end)
}

func Closing(market string, t time.Time) bool {
	_, end, ok := SessionBounds(market, t)
	return !ok || !t.Before(end.Add(-15*time.Minute))
}

func (ds Dataset) ClosingAt(t time.Time) bool {
	inst, _ := FindInstrument(ds.Symbol)
	if closeAt, ok := ds.SessionCloses[MarketDay(inst.Market, t)]; ok {
		return !t.Before(closeAt.Add(-15 * time.Minute))
	}
	return Closing(inst.Market, t)
}
