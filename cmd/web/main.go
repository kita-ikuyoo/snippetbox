package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/gob"
	"flag"
	"github.com/alexedwards/scs/v2"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"snippetbox/internal/models"
	"time"
)

type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	formDecoder    *form.Decoder
	users          *models.UserModel
}

func main() {
	port := flag.String("port", "443", "HTTP network port")
	// parseTime=true tells the driver to convert time type to golang's time.Time
	dsn := flag.String("dsn", "web:web@/snippetbox?parseTime=true&loc=Asia%2FTokyo", "MySQL data source name")
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
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	formDecoder := form.NewDecoder()

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		sessionManager: sessionManager,
		formDecoder:    formDecoder,
		users:          &models.UserModel{DB: db},
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	gob.Register(snippetCreateForm{})
	gob.Register(userSignupForm{})
	gob.Register(userLoginForm{})
	// http.Dir is FileSystem interface: Open(name string) (File, error)
	// http.FileServer returns handler, a interface: ServeHTTP(ResponseWriter, *Request)
	// log.Printf("%T\n", fileServer)
	// When a request comes in for /static/css/main.css,
	// the file server would look for it at ./ui/static/static/css/main.css —
	// the /static/ part gets duplicated.
	srv := http.Server{
		Addr:    ":" + *port,
		Handler: app.routes(),
		// http.ServerのErrorLogは*log.Logger
		// しかし、logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))は*log/slog.logger
		// slog.NewLogLogger の第二引数に渡すことで、「このロガーが出力するログを Error レベルとして扱う」という意味になります。
		ErrorLog:  slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		// keep-alive will be closed after 1 minute
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	app.logger.Info("starting server on", slog.String("port", *port))
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
