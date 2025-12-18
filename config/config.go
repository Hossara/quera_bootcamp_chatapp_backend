package config

import "fmt"

func (c *PostgresDBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		c.Host, c.Port, c.User, c.Pass, c.Name, c.SSLMode, c.Schema)
}
