package config

import (
	"os"
)

type Config struct {
	DeepSeekAPIKey string
	DeepSeekModel  string
	Port           string
	PersonaPath    string
	PromptPath     string
	AccessPassword string // 访问密码
}

func Load() *Config {
	return &Config{
		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:  getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		Port:           getEnv("PORT", "8080"),
		PersonaPath:    getEnv("PERSONA_PATH", "../data/persona.md"),
		PromptPath:     getEnv("PROMPT_PATH", "../data/prompt.md"),
		AccessPassword: getEnv("ACCESS_PASSWORD", ""), // 默认空字符串表示不启用密码
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
