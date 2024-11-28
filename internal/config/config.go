package config

// server
const (
	ListenScheme  string = "http"
	ListenAddress string = "0.0.0.0"
	ListenPort    string = "8080"
)

// agent
const (
	ServerScheme         = ListenScheme
	ServerAddress string = "localhost"
	ServerPort           = ListenPort
)
