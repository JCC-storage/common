package internal

import (
	"strings"
)

func makeEtcdLockRequestKey(reqID string) string {
	return EtcdLockRequestData + "/" + reqID
}

func getLockRequestID(key string) string {
	return strings.TrimPrefix(key, EtcdLockRequestData+"/")
}

/*
func parseLockData(str string) (lock lockData, err error) {
	sb := strings.Builder{}
	var comps []string

	escaping := false
	for _, ch := range strings.TrimSpace(str) {
		if escaping {
			if ch == 'n' {
				sb.WriteRune('\n')
			} else {
				sb.WriteRune(ch)
			}

			escaping = false
			continue
		}

		if ch == '/' {
			comps = append(comps, sb.String())
			sb.Reset()
		} else if ch == '\\' {
			escaping = true
		} else {
			sb.WriteRune(ch)
		}
	}

	comps = append(comps, sb.String())

	if len(comps) < 3 {
		return lockData{}, fmt.Errorf("string must includes 3 components devided by /")
	}

	return lockData{
		Path:   comps[0 : len(comps)-2],
		Name:   comps[len(comps)-2],
		Target: comps[len(comps)-1],
	}, nil
}

func lockDataToString(lock lockData) string {
	sb := strings.Builder{}

	for _, s := range lock.Path {
		sb.WriteString(lockDataEncoding(s))
		sb.WriteRune('/')
	}

	sb.WriteString(lockDataEncoding(lock.Name))
	sb.WriteRune('/')

	sb.WriteString(lockDataEncoding(lock.Target))

	return sb.String()
}

func lockDataEncoding(str string) string {
	sb := strings.Builder{}

	for _, ch := range str {
		if ch == '\\' {
			sb.WriteString("\\\\")
		} else if ch == '/' {
			sb.WriteString("\\/")
		} else if ch == '\n' {
			sb.WriteString("\\n")
		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}

func lockDataDecoding(str string) string {
	sb := strings.Builder{}

	escaping := false
	for _, ch := range str {
		if escaping {
			if ch == 'n' {
				sb.WriteRune('\n')
			} else {
				sb.WriteRune(ch)
			}

			escaping = false
			continue
		}

		if ch == '\\' {
			escaping = true

		} else {
			sb.WriteRune(ch)
		}
	}

	return sb.String()
}
*/
