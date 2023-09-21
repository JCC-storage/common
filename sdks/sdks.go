package sdks

import "fmt"

type CodeMessageError struct {
	Code    string
	Message string
}

func (e *CodeMessageError) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}
