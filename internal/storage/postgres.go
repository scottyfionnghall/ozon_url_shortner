package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	DB *sql.DB
}

// Функция создаёт новый объект sql.DB и возвращает ссылку на него, перед этим
// проверив доступность через пинг.
func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{DB: db}, nil
}

// Функция инициализирует базу данных, мигрируя таблицу в базу данных и возвращает
// ошибку если такова случилась при миграции.
func (s *PostgresStore) Init() error {
	return s.createTable()
}

func (s *PostgresStore) createTable() error {
	query := `CREATE TABLE IF NOT EXISTS url(
		id SERIAL PRIMARY KEY,
		domain VARCHAR(255),
		original_path VARCHAR(255),
		shorten_path VARCHAR(255)
	);
	`

	_, err := s.DB.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

// Функция добавляет URL в базу данных используя готовый объект URL как параметр.
func (s *PostgresStore) AddURL(ctx context.Context, url *URL) error {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
	}()

	_, err = tx.Exec("INSERT INTO url (domain, original_path,shorten_path) VALUES($1,$2,$3)",
		url.Domain, url.OriginalPath, url.ShortenPath)

	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

// Функция использует сокращённый URL как параметр, находит целый URL
// и возвращает ссылку на объект URL и ошибку если такова была.
func (s *PostgresStore) ReturnURL(ctx context.Context, short string) (*URL, error) {
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Rollback()
	}()

	rows, err := tx.Query("SELECT * FROM url WHERE shorten_path=$1", short)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoURL(rows)
	}

	return nil, fmt.Errorf("not found")
}

// Функция проверяет существует ли уже такой URL в базе данных, если существует
// то функция возвращает nil, если нет то функция возвращает уже существующий
// объект URL из базы данных.
func (s *PostgresStore) CheckExists(domain string, path string) (*URL, error) {

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		err = tx.Rollback()
	}()

	rows := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM url WHERE domain = $1 AND original_path = $2)", domain, path)

	var result bool

	err = rows.Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("error while checking if exists:%s", err.Error())
	}

	if result {
		rows, err := tx.Query("SELECT * FROM url WHERE domain = $1 AND original_path = $2", domain, path)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			return scanIntoURL(rows)
		}
	}

	err = tx.Commit()
	return nil, err
}

func scanIntoURL(rows *sql.Rows) (*URL, error) {
	url := new(URL)
	err := rows.Scan(
		&url.ID,
		&url.Domain,
		&url.OriginalPath,
		&url.ShortenPath,
	)

	return url, err
}

func (s *PostgresStore) ClearTable() error {

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback()
	}()
	_, err = tx.Exec("DROP TABLE url")
	return err
}
