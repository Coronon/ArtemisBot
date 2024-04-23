package config

type Config struct {
	// The Artemis base URL to use for HTTP requests
	ArtemisHttpURL string `json:"artemis_http_url"`
	// The Artemis base URL to use for WebSocket requests
	ArtemisWsURL string `json:"artemis_ws_url"`
}

var C *Config

func init() {
	C = &Config{
		ArtemisHttpURL: "https://artemis.in.tum.de/api",
		ArtemisWsURL:   "wss://artemis.in.tum.de/websocket",
	}
}
