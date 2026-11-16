package computers

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	"github.com/FacileStudio/tronc/errors"
)

// MaxNetAllow is the most extra hosts one computer may carry, the same cap the
// worker enforces.
const MaxNetAllow = 64

const maxHostLen = 253

// ValidateNetAllow checks the extra hosts of an allowlist computer with the
// worker's own rules, so a bad entry is answered here instead of failing at
// the worker. It lowercases, drops duplicates and returns the normalized list.
// An entry is host, *.host (subdomains only) or host:port. IP literals,
// single-label names, bare wildcards such as *.com and bad ports are refused.
func ValidateNetAllow(entries []string) ([]string, error) {
	if len(entries) > MaxNetAllow {
		return nil, errors.Invalid(fmt.Sprintf("net_allow holds at most %d hosts", MaxNetAllow))
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		rule, err := parseHostRule(entry)
		if err != nil {
			return nil, errors.Invalid("net_allow: " + err.Error())
		}
		if !seen[rule] {
			seen[rule] = true
			out = append(out, rule)
		}
	}
	return out, nil
}

func parseHostRule(entry string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(entry))
	if strings.HasPrefix(s, "[") || strings.Count(s, ":") > 1 {
		return "", fmt.Errorf("%q: IP literals are not allowed", entry)
	}
	port := 0
	if host, raw, ok := strings.Cut(s, ":"); ok {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 65535 {
			return "", fmt.Errorf("%q: invalid port", entry)
		}
		s, port = host, n
	}
	wildcard := false
	if rest, ok := strings.CutPrefix(s, "*."); ok {
		s, wildcard = rest, true
	}
	host, err := normalizeHost(s)
	if err != nil {
		return "", fmt.Errorf("%q: %w", entry, err)
	}
	if wildcard {
		host = "*." + host
	}
	if port != 0 {
		host += ":" + strconv.Itoa(port)
	}
	return host, nil
}

func normalizeHost(h string) (string, error) {
	h = strings.TrimSuffix(strings.ToLower(h), ".")
	if h == "" || len(h) > maxHostLen {
		return "", fmt.Errorf("invalid host length")
	}
	if _, err := netip.ParseAddr(h); err == nil {
		return "", fmt.Errorf("IP literals are not allowed")
	}
	labels := strings.Split(h, ".")
	if len(labels) < 2 {
		return "", fmt.Errorf("host needs at least two labels")
	}
	for _, l := range labels {
		if !validLabel(l) {
			return "", fmt.Errorf("invalid label %q", l)
		}
	}
	if tld := labels[len(labels)-1]; !strings.ContainsAny(tld, "abcdefghijklmnopqrstuvwxyz") || isHexNumber(tld) {
		return "", fmt.Errorf("top-level label %q looks like a number", tld)
	}
	return h, nil
}

func validLabel(l string) bool {
	if l == "" || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
		return false
	}
	for _, c := range l {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func isHexNumber(s string) bool {
	rest, ok := strings.CutPrefix(s, "0x")
	if !ok {
		return false
	}
	_, err := strconv.ParseUint("0"+rest, 16, 64)
	return err == nil
}
