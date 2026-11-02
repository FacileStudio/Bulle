package assistant

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

const maxSenderMap = 2000

func loadJSON(st State, key string, v any) {
	data, err := st.Load(key)
	if err != nil || data == nil {
		return
	}
	if err := json.Unmarshal(data, v); err != nil {
		return
	}
}

func updateJSON[T any](st State, key string, fn func(*T) bool) error {
	return st.Update(key, func(current []byte) ([]byte, error) {
		var v T
		if current != nil {
			if err := json.Unmarshal(current, &v); err != nil {
				var zero T
				v = zero
			}
		}
		if !fn(&v) {
			return nil, nil
		}
		return json.MarshalIndent(v, "", "  ")
	})
}

// RecordMailSenders remembers which address sent each mail key, so a 👎 on a
// digest line can find the sender to drop.
func RecordMailSenders(st State, senders map[string]string) error {
	if st == nil || len(senders) == 0 {
		return nil
	}
	return updateJSON(st, "mail-senders", func(m *map[string]string) bool {
		if *m == nil {
			*m = map[string]string{}
		}
		maps.Copy(*m, senders)
		for k := range *m {
			if len(*m) <= maxSenderMap {
				break
			}
			delete(*m, k)
		}
		return true
	})
}

// LoadDropSenders returns the senders added by 👎 reactions.
func LoadDropSenders(st State) []string {
	var list []string
	if st != nil {
		loadJSON(st, "drop-senders", &list)
	}
	return list
}

// DropSenderOf appends the sender address recorded for a mail key to
// the "drop-senders" state. An unknown key does nothing.
func DropSenderOf(st State, key string) error {
	m := map[string]string{}
	loadJSON(st, "mail-senders", &m)
	addr := strings.ToLower(strings.TrimSpace(m[key]))
	if addr == "" {
		return nil
	}
	return updateJSON(st, "drop-senders", func(list *[]string) bool {
		if slices.Contains(*list, addr) {
			return false
		}
		*list = append(*list, addr)
		return true
	})
}
