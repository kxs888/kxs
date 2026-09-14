package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	RoleAPI    = "api"
	RoleWorker = "worker"
	RoleAll    = "all"
)

// Config 仅从环境变量读取。国密密钥禁止入库、禁止写进仓库。
type Config struct {
	AppRole         string
	HTTPAddr        string
	HTTPTimeout     time.Duration
	HTTPBodyLimit   int64
	CORSOrigins     []string
	DatabaseURL     string
	JWTSecret       string
	JWTTTL          time.Duration
	LogLevel        string
	OTELEndpoint    string
	ServiceName     string
	SMCryptoEnabled bool
	SM4KeyHex       string
	BootstrapUser   string
	BootstrapPass   string
	ShutdownTimeout time.Duration
	OutboxInterval  time.Duration
}

func Load() (*Config, error) {
	_ = loadDotEnv(".env")
	c := &Config{
		AppRole:         envOr("APP_ROLE", RoleAll),
		HTTPAddr:        envOr("HTTP_ADDR", ":8080"),
		HTTPTimeout:     envDuration("HTTP_TIMEOUT", 15*time.Second),
		HTTPBodyLimit:   envBytes("HTTP_BODY_LIMIT", 1<<20),
		CORSOrigins:     splitCSV(envOr("CORS_ALLOWED_ORIGINS", "")),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTTTL:          envDuration("JWT_TTL", 15*time.Minute),
		LogLevel:        envOr("LOG_LEVEL", "info"),
		OTELEndpoint:    strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		ServiceName:     envOr("OTEL_SERVICE_NAME", "cga-api"),
		SMCryptoEnabled: envBool("SM_CRYPTO_ENABLED", false),
		SM4KeyHex:       strings.TrimSpace(os.Getenv("SM4_KEY_HEX")),
		BootstrapUser:   envOr("BOOTSTRAP_USERNAME", "demo"),
		BootstrapPass:   envOr("BOOTSTRAP_PASSWORD", "demo-pass-change-me"),
		ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		OutboxInterval:  envDuration("OUTBOX_INTERVAL", 2*time.Second),
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) Validate() error {
	switch c.AppRole {
	case RoleAPI, RoleWorker, RoleAll:
	default:
		return fmt.Errorf("APP_ROLE must be api|worker|all")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.JWTTTL <= 0 || c.JWTTTL > time.Hour {
		return fmt.Errorf("JWT_TTL must be a short access TTL (0 < ttl <= 1h)")
	}
	if c.SMCryptoEnabled && len(c.SM4KeyHex) != 32 {
		return fmt.Errorf("SM4_KEY_HEX must be 32 hex chars (16 bytes) when SM_CRYPTO_ENABLED=true")
	}
	return nil
}

func (c *Config) ServeHTTP() bool {
	return c.AppRole == RoleAPI || c.AppRole == RoleAll
}

func (c *Config) RunWorker() bool {
	return c.AppRole == RoleWorker || c.AppRole == RoleAll
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func envBytes(key string, def int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
	return sc.Err()
}
