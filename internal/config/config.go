package config

type Config struct {
	HTTPAddr    string
	DataBAseURL string
}

func MustLoad() Config {
	return Config{
		HTTPAddr:    ":8080",
		DataBAseURL: "postgres://wgtoooa:12354@127.0.0.1:5432/attendance?sslmode=disable",
	}
}
