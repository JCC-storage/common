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
			m.printer(fmt.Sprintf("[begin%v]%v:%v", titlePart, path.Base(file), line))
		} else {
			m.printer(fmt.Sprintf("[begin%v]unknown point", titlePart))
		}
	}
}

func (m *Measurement) Point(head ...string) {
	if m == nil {
		return
	}

	if m.on {
		m.printer(m.makePointString(strings.Join(head, ".")))
	}
}

func (m *Measurement) makePointString(head string) string {
	last := m.lastPointTime
	now := time.Now()
	m.lastPointTime = now

	_, file, line, ok := runtime.Caller(2)

	prefixCont := ""

	if m.title != "" {
		prefixCont = m.title
	}

	if head != "" {
		if prefixCont == "" {
			prefixCont = head
		} else {
			prefixCont = fmt.Sprintf("%s.%s", prefixCont, head)
		}
	}

	prefixPart := ""
	if prefixCont != "" {
		prefixPart = fmt.Sprintf("[%s]", prefixCont)
	}

	if ok {
		return fmt.Sprintf("%v%v:%v@%v(%v)", prefixPart, path.Base(file), line, now.Sub(last), now.Sub(m.startTime))
	}

	return fmt.Sprintf("%vunknown point@%v(%v)", prefixPart, now.Sub(last), now.Sub(m.startTime))
}

func (m *Measurement) End(head ...string) {
	if m == nil {
		return
	}

	if m.on {
		m.printer(fmt.Sprintf("[end]%v\n", m.makePointString(strings.Join(head, "."))))
	}
}

