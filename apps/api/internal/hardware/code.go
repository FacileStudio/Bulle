package hardware

import (
	"crypto/rand"
	"strings"
)

const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const codeLength = 8

// NormalizeCode uppercases a typed code and drops spaces and dashes.
func NormalizeCode(code string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' || r == ' ' {
			return -1
		}
		return r
	}, strings.ToUpper(strings.TrimSpace(code)))
}

// FormatCode splits a user code in two groups for reading aloud.
func FormatCode(code string) string {
	return code[:codeLength/2] + "-" + code[codeLength/2:]
}

func newUserCode() string {
	buf := make([]byte, codeLength)
	out := make([]byte, codeLength)
	for i := 0; i < codeLength; {
		_, _ = rand.Read(buf)
		for _, b := range buf {
			if i == codeLength {
				break
			}
			if int(b) < 256-256%len(codeAlphabet) {
				out[i] = codeAlphabet[int(b)%len(codeAlphabet)]
				i++
			}
		}
	}
	return string(out)
}
