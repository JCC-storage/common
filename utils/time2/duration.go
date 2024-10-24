package time2

import (
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) Std() time.Duration {
	return d.Duration
}

func (d *Duration) Scan(state fmt.ScanState, verb rune) error {
	data, err := state.Token(true, nil)
	if err != nil {
		return err
	}

	d.Duration, err = time.ParseDuration(string(data))
	if err != nil {
		return err
	}

	return nil
}
