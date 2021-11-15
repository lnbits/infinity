package utils

import "time"

func DateStringToTimestamp(datestr string) (int64, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", datestr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}
