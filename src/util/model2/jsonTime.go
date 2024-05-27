package model2

import "time"

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(j).Format(time.DateTime) + `"`), nil
}

func String2JsonTime(str string) JsonTime {
	t, _ := time.Parse(time.DateTime, str)
	return JsonTime(t)
}
