package config

type ServerConfig struct {
	ServerAddr string // flag -a
	BaseURL    string // flag -b
}

func NewServerConfig(serverAddr, baseURL string) ServerConfig {
	return ServerConfig{
		ServerAddr: serverAddr,
		BaseURL:    baseURL,
	}
}
