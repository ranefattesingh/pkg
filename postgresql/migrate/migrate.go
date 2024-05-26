package migrate

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ranefattesingh/pkg/postgresql"
)

type PostgresMigrator struct {
	migrate *migrate.Migrate
}

func NewDatabaseMigrator(migrationDir string, connectionString string) (*PostgresMigrator, error) {
	migrate, err := migrate.New("file://"+migrationDir, postgresql.EncodeConnectionString(connectionString))
	if err != nil {
		return nil, err
	}

	return &PostgresMigrator{migrate: migrate}, nil
}

func (p *PostgresMigrator) Migrate() *migrate.Migrate {
	return p.migrate
}

func (p *PostgresMigrator) CloseDatabaseMigrator() (source error, database error) {
	return p.migrate.Close()
}

func (p *PostgresMigrator) GracefulStop() {
	p.migrate.GracefulStop <- true
}
