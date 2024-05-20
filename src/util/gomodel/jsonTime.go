package model

import "time"

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(j).Format(time.DateTime) + `"`), nil
}
