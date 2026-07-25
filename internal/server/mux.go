package server

import (
	"database/sql"
	"net/http"

	"github.com/velocity-center-makerspace/maintenance-tracker/db"
	"github.com/velocity-center-makerspace/maintenance-tracker/internal/config"
)

type MuxParams struct {
	DB  *sql.DB
	Qry *db.Queries
	Env *config.Environment
}

type Mux struct {
	*http.ServeMux
	DB  *sql.DB
	Qry *db.Queries
	Env *config.Environment
}

func NewMux(args MuxParams) *Mux {
	mux := &Mux{
		ServeMux: http.NewServeMux(),
		DB:       args.DB,
		Qry:      args.Qry,
		Env:      args.Env,
	}

	return mux
}
