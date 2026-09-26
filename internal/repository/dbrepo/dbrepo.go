package dbrepo

import (
	"database/sql"

	"github.com/GangaRamPrasad2004/sentinal/internal/config"
	"github.com/GangaRamPrasad2004/sentinal/internal/repository"
	"github.com/GangaRamPrasad2004/sentinal/internal/repository/dbrepo/sqlc"
)

var app *config.AppConfig

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
	q   *sqlc.Queries
}

// NewPostgresRepo creates the repository
func NewPostgresRepo(Conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	app = a
	return &postgresDBRepo{
		App: a,
		DB:  Conn,
		q:   sqlc.New(Conn),
	}
}

type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewTestingRepo creates a repo with a dummy database for testing
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testDBRepo{
		App: a,
	}
}
