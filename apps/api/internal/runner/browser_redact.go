package runner

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

const redactedValue = "[redacted]"

var (
	jsonPair       = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"(\s*:\s*)("(?:[^"\\]|\\.)*"?|[^,{}\[\]\s]+)`)
	longSecretWord = []string{"password", "passwd", "token", "secret", "apikey", "card"}
	shortSecret    = map[string]bool{"pin": true, "otp": true, "cvv": true, "cvc": true}
)

// redactBody hides the values of secret-looking fields in a held request's
// body preview, JSON or form-encoded, so the approval card shows what is
// being sent without showing a password or a card number.
func redactBody(body string, redact Redactor) string {
	trimmed := strings.TrimSpace(body)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		body = jsonPair.ReplaceAllStringFunc(body, redactJSONPair)
	} else {
		body = redactForm(body)
	}
	return redact.Apply(body)
}

func redactJSONPair(pair string) string {
	parts := jsonPair.FindStringSubmatch(pair)
	if !secretField(parts[1]) {
		return pair
	}
	return `"` + parts[1] + `"` + parts[2] + `"` + redactedValue + `"`
}

func redactForm(body string) string {
	fields := strings.Split(body, "&")
	for i, field := range fields {
		name, _, found := strings.Cut(field, "=")
		decoded, err := url.QueryUnescape(name)
		if found && err == nil && secretField(decoded) {
			fields[i] = name + "=" + redactedValue
		}
	}
	return strings.Join(fields, "&")
}

// secretField reports whether a field name looks like it holds a secret. The
// name is split into words (snake, kebab, bracket and camel case), and a word
// counts when it is pin, otp, cvv or cvc, or starts or ends with one of the
// longer secret words such as password or card.
func secretField(name string) bool {
	words := fieldWords(name)
	if strings.Contains(strings.Join(words, ""), "apikey") {
		return true
	}
	for _, word := range words {
		if shortSecret[word] {
			return true
		}
		for _, secret := range longSecretWord {
			if strings.HasPrefix(word, secret) || strings.HasSuffix(word, secret) {
				return true
			}
		}
	}
	return false
}

func fieldWords(name string) []string {
	var spaced strings.Builder
	var previous rune
	for _, r := range name {
		if unicode.IsUpper(r) && unicode.IsLower(previous) {
			spaced.WriteRune(' ')
		}
		spaced.WriteRune(unicode.ToLower(r))
		previous = r
	}
	return strings.FieldsFunc(spaced.String(), func(r rune) bool { return !unicode.IsLetter(r) })
}
