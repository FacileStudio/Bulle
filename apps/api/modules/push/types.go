package push

type StatusResponse struct {
	Enabled      bool   `json:"enabled"`
	Server       string `json:"server"`
	Topic        string `json:"topic"`
	SubscribeURL string `json:"subscribe_url"`
}

type EmptyRequest struct{}

type EmptyResponse struct{}
