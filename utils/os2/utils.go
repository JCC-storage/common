package os2

import (
	"math/rand"
	"strings"
)

func GenerateRandomFileName(len int) string {
	sb := strings.Builder{}
	for i := 0; i < len; i++ {
		rd := rand.Intn(26 + 10)
		if rd < 26 {
			sb.WriteRune('a' + rune(rd))
		} else {
			sb.WriteRune('0' + rune(rd-10))
		}
	}

	return sb.String()
}
