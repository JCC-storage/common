package serder

import (
	"fmt"
	"strconv"
	"time"
)

type TimestampSecond time.Time

func (t *TimestampSecond) MarshalJSON() ([]byte, error) {
	raw := time.Time(*t)
	return []byte(fmt.Sprintf("%d", raw.Unix())), nil
}

func (t *TimestampSecond) UnmarshalJSON(data []byte) error {
	var timestamp int64
	var err error
	if timestamp, err = strconv.ParseInt(string(data), 10, 64); err != nil {
		return err
	}

	*t = TimestampSecond(time.Unix(timestamp, 0))
	return nil
}

type TimestampMilliSecond time.Time

func (t *TimestampMilliSecond) MarshalJSON() ([]byte, error) {
	raw := time.Time(*t)
	return []byte(fmt.Sprintf("%d", raw.UnixMilli())), nil
}

func (t *TimestampMilliSecond) UnmarshalJSON(data []byte) error {
	var timestamp int64
	var err error
	if timestamp, err = strconv.ParseInt(string(data), 10, 64); err != nil {
		return err
	}

	*t = TimestampMilliSecond(time.UnixMilli(timestamp))
	return nil
}
