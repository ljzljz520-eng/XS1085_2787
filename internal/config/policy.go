package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type RuntimePolicy struct {
	MaxBodyBytes   int64         `json:"max_body_bytes"`
	AnimationFrame time.Duration `json:"animation_frame"`
	SparkLimit     int           `json:"spark_limit"`
	SessionTTL     time.Duration `json:"session_ttl"`
	AllowedOrigins []string      `json:"allowed_origins"`
}

func DefaultPolicy() RuntimePolicy {
	return RuntimePolicy{MaxBodyBytes: 1 << 20, AnimationFrame: 12 * time.Millisecond, SparkLimit: 24, SessionTTL: 30 * time.Minute, AllowedOrigins: []string{"*"}}
}

func (p RuntimePolicy) Valid() bool {
	return p.MaxBodyBytes >= 1024 && p.MaxBodyBytes <= 16<<20 && p.AnimationFrame > 0 && p.AnimationFrame <= time.Second && p.SparkLimit >= 1 && p.SparkLimit <= 500 && p.SessionTTL >= time.Minute && len(p.AllowedOrigins) > 0
}

func (p RuntimePolicy) Normalize() RuntimePolicy {
	if p.MaxBodyBytes < 1024 {
		p.MaxBodyBytes = 1024
	}
	if p.MaxBodyBytes > 16<<20 {
		p.MaxBodyBytes = 16 << 20
	}
	if p.AnimationFrame <= 0 {
		p.AnimationFrame = 12 * time.Millisecond
	}
	if p.SparkLimit < 1 {
		p.SparkLimit = 1
	}
	if p.SparkLimit > 500 {
		p.SparkLimit = 500
	}
	if p.SessionTTL < time.Minute {
		p.SessionTTL = time.Minute
	}
	if len(p.AllowedOrigins) == 0 {
		p.AllowedOrigins = []string{"*"}
	}
	return p
}

func PolicyFromEnv() RuntimePolicy {
	policy := DefaultPolicy()
	if value := os.Getenv("MEMORIAL_MAX_BODY"); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			policy.MaxBodyBytes = parsed
		}
	}
	if value := os.Getenv("MEMORIAL_FRAME"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			policy.AnimationFrame = parsed
		}
	}
	if value := os.Getenv("MEMORIAL_SPARK_LIMIT"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			policy.SparkLimit = parsed
		}
	}
	if value := os.Getenv("MEMORIAL_SESSION_TTL"); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			policy.SessionTTL = parsed
		}
	}
	if value := os.Getenv("MEMORIAL_ORIGINS"); value != "" {
		policy.AllowedOrigins = splitList(value)
	}
	return policy.Normalize()
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func ValidateAddress(address string) error {
	if strings.TrimSpace(address) == "" {
		return errors.New("address is required")
	}
	if strings.HasPrefix(address, "/") {
		return errors.New("address must include a port")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("address is invalid: %w", err)
	}
	if host == "" {
		host = "0.0.0.0"
	}
	if host != "0.0.0.0" && net.ParseIP(host) == nil && strings.Contains(host, " ") {
		return errors.New("address host contains spaces")
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func ValidateDatabase(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("database path is required")
	}
	if strings.HasPrefix(path, "file:") {
		return nil
	}
	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".db" && extension != ".sqlite" && extension != ".sqlite3" {
		return errors.New("database path must use a sqlite extension")
	}
	if filepath.IsAbs(path) {
		return errors.New("database path must be relative")
	}
	return nil
}

func (c Config) ValidateAll(policy RuntimePolicy) error {
	if !c.Valid() {
		return errors.New("base configuration is invalid")
	}
	if err := ValidateAddress(c.Addr); err != nil {
		return err
	}
	if err := ValidateDatabase(c.Database); err != nil {
		return err
	}
	if !policy.Valid() {
		return errors.New("runtime policy is invalid")
	}
	return nil
}

func Describe(c Config, p RuntimePolicy) string {
	return fmt.Sprintf("addr=%s database=%s seed=%s frame=%s sparks=%d", c.Addr, c.Database, c.Seed, p.AnimationFrame, p.SparkLimit)
}
