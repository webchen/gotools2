package model2

import "time"

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(j).Format(time.DateTime) + `"`), nil
}

func (j JsonTime) ToString() string {
	return time.Time(j).Format(time.DateTime)
}

func String2JsonTime(str string) JsonTime {
	t, _ := time.ParseInLocation(time.DateTime, str, time.Local)
	return JsonTime(t)
}
