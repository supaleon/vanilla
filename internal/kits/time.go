package kits

import (
	"strings"
	"time"
)

const (
	TimeNormal = "2006-01-02 15:04:05"
	TimeShort  = "20060102150405"
)

type Time struct {
	time.Time
}

func NewTime(t time.Time) *Time {
	return &Time{Time: t}
}

func (t *Time) String() string {
	return t.Format(TimeNormal)
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Format(TimeNormal) + `"`), nil
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	ts := strings.Trim(string(data), "\"")
	var tm time.Time
	// 2006-01-02T15:04:05Z07:00
	if strings.Contains(ts, "T") {
		tm, err = time.Parse(time.RFC3339, ts)
		if err == nil {
			t.Time = tm
			return
		}
	}
	// 20060102150405
	if !strings.Contains(ts, " ") {
		tm, err = time.Parse(TimeShort, ts)
		if err == nil {
			t.Time = tm
			return
		}
	}
	// 2006-01-02 15:04:05 (as java do.)
	tm, err = time.Parse(TimeNormal, ts)
	if err == nil {
		t.Time = tm
	}
	return
}
