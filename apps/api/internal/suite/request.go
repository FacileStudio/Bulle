package suite

import "encoding/json"

type Request struct {
	App         string
	Method      string
	Path        string
	Body        any
	Raw         []byte
	ContentType string
}

func (req Request) encode() ([]byte, string, error) {
	if req.Raw != nil {
		return req.Raw, req.ContentType, nil
	}
	if req.Body == nil {
		return nil, "", nil
	}
	encoded, err := json.Marshal(req.Body)
	return encoded, "application/json", err
}

func clampLimit(limit, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	return min(limit, 50)
}
