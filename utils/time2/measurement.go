package time2

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"
)

type Measurement struct {
	startTime     time.Time
	lastPointTime time.Time
	printer       func(string)
	on            bool
	title         string
}

func NewMeasurement(printer func(string)) Measurement {
	return Measurement{
		printer: printer,
	}
}

func (m *Measurement) Begin(on bool, title ...string) {
	if m == nil {
		return
	}

	m.on = on
	m.title = strings.Join(title, ".")

	if on {
		m.startTime = time.Now()
		m.lastPointTime = m.startTime

		_, file, line, ok := runtime.Caller(1)

		titlePart := ""
		if m.title != "" {
			titlePart = fmt.Sprintf(":%s", m.title)
		}

		if ok {
			m.printer(fmt.Sprintf("[BEGIN%v]<%v:%v>", titlePart, path.Base(file), line))
		} else {
			m.printer(fmt.Sprintf("[BEGIN%v]<UNKNOWN>", titlePart))
		}
	}
}

func (m *Measurement) Point(desc ...string) {
	if m == nil {
		return
	}

	if m.on {
		m.printer(m.makePointString(strings.Join(desc, ".")))
	}
}

func (m *Measurement) makePointString(desc string) string {
	last := m.lastPointTime
	now := time.Now()
	m.lastPointTime = now

	_, file, line, ok := runtime.Caller(2)

	titlePart := ""
	if m.title != "" {
		titlePart = fmt.Sprintf("(%s)", m.title)
	}

	if desc != "" {
		desc = fmt.Sprintf("@%s", desc)
	}

	if ok {
		return fmt.Sprintf("%v {%v/%v} %v<%v:%v>", titlePart, now.Sub(last), now.Sub(m.startTime), desc, path.Base(file), line)
	}

	return fmt.Sprintf("{%v/%v}%v<UNKNOWN>", now.Sub(last), now.Sub(m.startTime), desc)
}

func (m *Measurement) End(descs ...string) {
	if m == nil {
		return
	}

	if m.on {
		last := m.lastPointTime
		now := time.Now()
		m.lastPointTime = now

		_, file, line, ok := runtime.Caller(1)

		titlePart := ""
		if m.title != "" {
			titlePart = fmt.Sprintf(":%s", m.title)
		}

		desc := strings.Join(descs, ".")
		if desc != "" {
			desc = fmt.Sprintf("@%s", desc)
		}

		if ok {
			m.printer(fmt.Sprintf("[END%v] {%v/%v} %v<%v:%v>", titlePart, now.Sub(last), now.Sub(m.startTime), desc, path.Base(file), line))
		} else {
			m.printer(fmt.Sprintf("[END%v] {%v/%v} %v<UNKNOWN>", titlePart, now.Sub(last), now.Sub(m.startTime), desc))
		}
	}
}
