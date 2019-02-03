package database

import (
	"database/sql"

	"github.com/nburunova/taxi-backend-sample/src/infrastructure/log"

	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

var maxAttempts = 10

// Database - структура для доступа к базе данных
type Database struct {
	db     *sql.DB
	logger log.Logger
}

// New - создаем новую структуру для работы с базой данных
func New(dsn string, logger log.Logger) (*Database, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Prostgres DB: could not open connection")
	}
	return &Database{
		db:     db,
		logger: logger,
	}, nil
}

// Connect - присоединяемся к базе
func (d *Database) Connect() error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = d.db.Ping()
		if err == nil {
			return nil
		}
		nextAttemptWait := time.Duration(attempt) * time.Second
		d.logger.Warnf("attempt %v: could not establish a connection with the database, wait for %v.", attempt, nextAttemptWait)
		time.Sleep(nextAttemptWait)
	}
	if err != nil {
		return errors.Wrap(err, "Prostgres DB: could not connect to database")
	}
	return nil
}

// Close - закрываем соединение к базе
func (d *Database) Close() error {
	if err := d.db.Close(); err != nil {
		return errors.Wrap(err, "Prostgres DB: could not close database")
	}
	return nil
}

// Query - запрос к базе
func (d *Database) Query(queryString string) (*sql.Rows, error) {
	d.logger.Debugf("Request DB %v", queryString)
	return d.db.Query(queryString)
}
