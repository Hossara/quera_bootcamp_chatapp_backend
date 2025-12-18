package database

import (
	"context"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/config"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
)

func NewEntClient(cfg config.PostgresDBConfig) (*ent.Client, error) {
	db, err := openSQLDB(cfg)
	if err != nil {
		return nil, err
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	return client, nil
}

func MigrateEnt(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(ctx)
}
