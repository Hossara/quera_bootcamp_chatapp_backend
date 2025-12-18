package config

// DefaultConfig holds the default values for the configuration.
var DefaultConfig = Config{
	Server: ServerConfig{
		Port:        8080,
		Host:        "0.0.0.0",
		Environment: "development",
		TestToken:   "",
		ProxySetting: ProxySettingConfig{
			Enabled: false,
			Host:    "127.0.0.1",
			Port:    2080,
			Type:    "socks5",
		},
	},
	PostgresDB: PostgresDBConfig{
		Host:    "localhost",
		Port:    5432,
		User:    "pguser",
		Pass:    "password",
		Name:    "chatapp",
		Schema:  "public",
		SSLMode: "disable",
	},
	Redis: RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
	},
	Logger: LoggerConfig{
		LogLevel: "info",
	},
	Auth: AuthConfig{
		PasetoKey:       "your-32-character-secret-key!",
		TokenExpiration: 24, // 24 hours
	},
}
