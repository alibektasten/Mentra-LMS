package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	AI          AIConfig
	Ollama      OllamaConfig
	OpenAI      OpenAIConfig
	GoogleOAuth GoogleOAuthConfig
	SuperAdmin  SuperAdminConfig
}

type SuperAdminConfig struct {
	Email    string
	Password string
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	DSN                    string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

type JWTConfig struct {
	Secret                string
	AccessTokenTTLMinutes int
}

type OllamaConfig struct {
	BaseURL        string
	Model          string
	TimeoutSeconds int
}

// AIConfig selects the implementation used for all AI-backed LMS features.
// Valid values are "ollama" and "openai".
type AIConfig struct {
	Provider          string
	DailyRequestLimit int
}

type OpenAIConfig struct {
	APIKey          string
	BaseURL         string
	Model           string
	TimeoutSeconds  int
	MaxOutputTokens int
}

// GoogleOAuthConfig contains the server-only credentials for Google sign-in.
// Leave ClientID and ClientSecret empty to keep the provider disabled.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	FrontendURL  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			DSN:                    mustGetEnv("DB_DSN"),
			MaxOpenConns:           getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeMinutes: getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30),
		},
		JWT: JWTConfig{
			Secret:                mustGetEnv("JWT_SECRET"),
			AccessTokenTTLMinutes: getEnvInt("JWT_ACCESS_TOKEN_TTL_MINUTES", 15),
		},
		AI: AIConfig{
			Provider:          getEnv("AI_PROVIDER", "ollama"),
			DailyRequestLimit: getEnvInt("AI_DAILY_REQUEST_LIMIT", 50),
		},
		Ollama: OllamaConfig{
			BaseURL:        getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:          getEnv("OLLAMA_MODEL", "llama3.2"),
			TimeoutSeconds: getEnvInt("OLLAMA_TIMEOUT_SECONDS", 60),
		},
		OpenAI: OpenAIConfig{
			APIKey:          os.Getenv("OPENAI_API_KEY"),
			BaseURL:         getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			Model:           getEnv("OPENAI_MODEL", "gpt-5.4-mini"),
			TimeoutSeconds:  getEnvInt("OPENAI_TIMEOUT_SECONDS", 60),
			MaxOutputTokens: getEnvInt("OPENAI_MAX_OUTPUT_TOKENS", 1200),
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
			RedirectURL:  getEnv("GOOGLE_OAUTH_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
			FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		SuperAdmin: SuperAdminConfig{
			Email:    getEnv("SUPER_ADMIN_EMAIL", ""),
			Password: getEnv("SUPER_ADMIN_PASSWORD", ""),
		},
	}
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
