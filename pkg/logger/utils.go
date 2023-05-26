package logger

import (
	"fmt"
	"reflect"
	"strings"
)

type structFormatter struct {
	val any
}

func (f *structFormatter) String() string {
	typ := reflect.TypeOf(f.val)
	val := reflect.ValueOf(f.val)

	kind := typ.Kind()

	if kind != reflect.Struct {
		return fmt.Sprintf("%v", f.val)
	}

	strBuilder := strings.Builder{}
	for i := 0; i < val.NumField(); i++ {
		fieldInfo := typ.Field(i)
		fieldValue := val.Field(i)
		fieldType := fieldInfo.Type
		fieldKind := fieldType.Kind()

		switch fieldKind {
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			if i > 0 {
				strBuilder.WriteString(", ")
			}

			strBuilder.WriteString("len(")
			strBuilder.WriteString(fieldInfo.Name)
			strBuilder.WriteString("): ")
			strBuilder.WriteString(fmt.Sprintf("%d", fieldValue.Len()))

		case reflect.Struct:
			if i > 0 {
				strBuilder.WriteString(", ")
			}

			strBuilder.WriteString(fieldInfo.Name)
			strBuilder.WriteString(": <")
			strBuilder.WriteString(fieldType.Name())
			strBuilder.WriteString(">")

		case reflect.Pointer:
			if i > 0 {
				strBuilder.WriteString(", ")
			}
			strBuilder.WriteString(fieldInfo.Name)
			if fieldValue.IsNil() {
				strBuilder.WriteString(": <nil>")
			} else {
				strBuilder.WriteString(": &<")
				strBuilder.WriteString(fieldType.Elem().Name())
				strBuilder.WriteString(">")
			}

		default:
			if i > 0 {
				strBuilder.WriteString(", ")
			}
			strBuilder.WriteString(fieldInfo.Name)
			strBuilder.WriteString(": ")
			strBuilder.WriteString(fmt.Sprintf("%v", fieldValue))
		}
	}

	return strBuilder.String()
}

func FormatStruct(val any) any {
	return &structFormatter{
		val: val,
	}
}
