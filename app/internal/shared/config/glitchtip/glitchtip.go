package glitchtip

import "os"

type Config struct {
	DSN string

	FrontendDSN string
}

func Read() Config {
	return Config{DSN: os.Getenv("GLITCHTIP_DSN"), FrontendDSN: os.Getenv("GLITCHTIP_FRONTEND_DSN")}
}
