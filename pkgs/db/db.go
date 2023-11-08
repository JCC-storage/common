package db

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
)

type IntString string

func (j IntString) Value() (driver.Value, error) {
	return strconv.ParseInt(string(j), 10, 64)
}

func (j *IntString) Scan(src interface{}) error {
	if src == nil {
		return fmt.Errorf("cannot convert nil to string")
	}

	bs, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %s to string", reflect.TypeOf(src).String())
	}

	var v int64
	if len(bs) == 8 {
		v = int64(binary.LittleEndian.Uint64(bs))
	} else if len(bs) == 4 {
		v = int64(binary.LittleEndian.Uint32(bs))
	} else {
		return fmt.Errorf("invalid bytes array length %d", len(bs))
	}

	*j = IntString(fmt.Sprintf("%d", v))

	return nil
}
