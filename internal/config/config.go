package config

type Config struct {
	HTTPAddr    string
	DataBAseURL string
}

func MustLoad() Config {
	return Config{
		HTTPAddr:    ":8080",
		DataBAseURL: "postgres://wgtoooa:12354@localhost:5432/ettendance?sslmode=disable",
	}
}
