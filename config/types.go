package config

// Config represents the configuration structure for the server.
type Config struct {
	Server     ServerConfig     `mapstructure:"general"`
	PostgresDB PostgresDBConfig `mapstructure:"postgres_db"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Auth       AuthConfig       `mapstructure:"auth"`
}

// AuthConfig represents the authentication configuration structure.
type AuthConfig struct {
	PasetoKey       string `mapstructure:"paseto_key"`
	TokenExpiration int    `mapstructure:"token_expiration"` // in hours
}

// ServerConfig represents the general server configuration structure.
type ServerConfig struct {
	Port         uint               `mapstructure:"port"`
	Host         string             `mapstructure:"host"`
	TestToken    string             `mapstructure:"test_token"`
	Environment  string             `mapstructure:"environment"`
	ProxySetting ProxySettingConfig `mapstructure:"proxy_setting"`
}

// ProxySettingConfig represents the proxy settings configuration structure.
type ProxySettingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    uint   `mapstructure:"port"`
	Type    string `mapstructure:"type"`
}

// PostgresDBConfig represents the configuration structure for PostgreSQL database.
type PostgresDBConfig struct {
	Host    string `mapstructure:"host"`
	Port    uint   `mapstructure:"port"`
	User    string `mapstructure:"user"`
	Pass    string `mapstructure:"pass"`
	Name    string `mapstructure:"name"`
	Schema  string `mapstructure:"schema"`
	SSLMode string `mapstructure:"ssl_mode"`
}

// LoggerConfig represents the configuration structure for the logger.
type LoggerConfig struct {
	LogLevel string `mapstructure:"log_level"`
}

// RedisConfig represents the configuration structure for Redis.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint   `mapstructure:"port"`
	Password string `mapstructure:"password"`
}
