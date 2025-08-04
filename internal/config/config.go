package config

type Config struct {
	SystemPort string `json:"system_port"`
	ListenPort string `json:"listen_port"`

	ServerOpts ServerOpts `json:"server_opts"`
}

func ParseConfig() (*Config, error) {
	return &Config{
		SystemPort: ":8081",
		ListenPort: ":8080",
		ServerOpts: ServerOpts{
			ReadTimeout:          60,
			WriteTimeout:         60,
			IdleTimeout:          60,
			MaxRequestBodySizeMb: 1,
		},
	}, nil
}

func (c *Config) Validate() error {
	return nil
}

type ServerOpts struct {
	ReadTimeout          int `json:"read_timeout"`
	WriteTimeout         int `json:"write_timeout"`
	IdleTimeout          int `json:"idle_timeout"`
	MaxRequestBodySizeMb int `json:"max_request_body_size_mb"`
}
