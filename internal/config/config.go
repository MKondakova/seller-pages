package config

type Config struct {
	ListenPort string

	ServerOpts    ServerOpts
	ProductsPath  string
	FeedbacksPath string
}

func GetConfig() (*Config, error) {
	return &Config{
		ListenPort: ":8080",
		ServerOpts: ServerOpts{
			ReadTimeout:          60,
			WriteTimeout:         60,
			IdleTimeout:          60,
			MaxRequestBodySizeMb: 1,
		},
		ProductsPath: "data/enriched_products.json",
	}, nil
}

type ServerOpts struct {
	ReadTimeout          int `json:"read_timeout"`
	WriteTimeout         int `json:"write_timeout"`
	IdleTimeout          int `json:"idle_timeout"`
	MaxRequestBodySizeMb int `json:"max_request_body_size_mb"`
}
