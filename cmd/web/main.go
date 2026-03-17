package main

import (
	"database/sql"
	"encoding/gob"
	"flag"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"snippetbox/internal/models"
)

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
}

func main() {
	port := flag.String("port", "4000", "HTTP network port")
	// parseTime=true tells the driver to convert time type to golang's time.Time
	dsn := flag.String("dsn", "web:web@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error(), slog.String("trace", string(debug.Stack())))
		os.Exit(1)
	}
	defer db.Close()
	// AddSource adds the source of error into output log
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error(), slog.String("trace", string(debug.Stack())))
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Store = memstore.New()

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
	}
	gob.Register(snippetCreateForm{})
	// http.Dir is FileSystem interface: Open(name string) (File, error)
	// http.FileServer returns handler, a interface: ServeHTTP(ResponseWriter, *Request)
	// log.Printf("%T\n", fileServer)
	// When a request comes in for /static/css/main.css,
	// the file server would look for it at ./ui/static/static/css/main.css —
	// the /static/ part gets duplicated.

	app.logger.Info("starting server on", slog.String("port", *port))
	err = http.ListenAndServe(":"+*port, app.routes())
	app.logger.Error(err.Error(), slog.String("trace", string(debug.Stack())))
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, err
}
