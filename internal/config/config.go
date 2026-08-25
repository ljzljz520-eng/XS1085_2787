package config

import "os"

type Config struct {
	Addr       string
	Database   string
	Seed       string
	QuietFrame int
}

func Default() Config {
	return Config{Addr: ":8080", Database: "memorial.db", Seed: "fixed-memorial-seed", QuietFrame: 120}
}

func FromEnv() Config {
	c := Default()
	if value := os.Getenv("MEMORIAL_ADDR"); value != "" {
		c.Addr = value
	}
	if value := os.Getenv("MEMORIAL_DB"); value != "" {
		c.Database = value
	}
	if value := os.Getenv("MEMORIAL_SEED"); value != "" {
		c.Seed = value
	}
	return c
}

func (c Config) Valid() bool {
	return c.Addr != "" && c.Database != "" && c.Seed != "" && c.QuietFrame > 0
}
