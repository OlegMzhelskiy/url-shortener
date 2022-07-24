package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"time"
	//_ "github.com/lib/pq"
	//"database/sql"
)

var (
	ErrDBConnection = errors.New("you haven`t opened the database connection")
	ErrURLDeleted   = errors.New("item has been deleted")
)

type StoreDB struct {
	db          *sqlx.DB
	config      *StoreConfig
	insertStmt  *sql.Stmt
	makeDelStmt *sql.Stmt
}

type RowTabURL struct {
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"origin_url"`
	UserID      string `db:"user_id"`
	Deleted     bool   `db:"deleted"`
}

func NewStoreDB(config *StoreConfig) (*StoreDB, error) {
	db, err := sqlx.Open("postgres", config.DBDNS)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls(
		origin_url TEXT UNIQUE NOT NULL,
		short_url TEXT NOT NULL,
		user_id TEXT NOT NULL,
        deleted boolean NOT NULL DEFAULT false);
		CREATE TABLE IF NOT EXISTS users_token(		
		user_id TEXT UNIQUE NOT NULL);`)
	//id SERIAL PRIMARY KEY
	//userId INTEGER REFERENCES users_token (id),

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(time.Second * 30)
	db.SetConnMaxLifetime(time.Minute * 2)

	insertStmt, err := db.Prepare("INSERT INTO urls(user_id, origin_url, short_url) VALUES($1,$2,$3)")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	makeDelStmt, err := db.Prepare("UPDATE urls SET deleted = true WHERE short_url = $1")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &StoreDB{db, config, insertStmt, makeDelStmt}, nil
}

func (store *StoreDB) Close() {
	if store.db != nil {
		store.db.Close()
	}
}

func (store *StoreDB) Ping() bool {
	if err := store.db.Ping(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (store StoreDB) GetAll(ctx context.Context) map[string]UserURL {
	urls := make(map[string]UserURL)
	rows, err := store.db.QueryContext(ctx, "SELECT user_id, origin_url, short_url, deleted FROM urls")

	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var (
				us  string
				or  string
				sh  string
				del bool
			)
			err = rows.Scan(&us, &or, &sh, &del)
			if err != nil {
				return urls
			}
			urls[sh] = UserURL{or, us, del}
		}
		err = rows.Err()
		if err != nil {
			return urls
		}
	}
	return urls
}

func (store StoreDB) SaveLink(ctx context.Context, shortURL, longURL, userID string) error {
	if store.db == nil {
		return ErrDBConnection //errors.New("you haven`t opened the database connection")
	}
	//url, _ := store.GetByID(shortURL)
	//Если записи с таким url нет то добавим
	//if url == "" {
	_, err := store.db.ExecContext(ctx, "INSERT INTO urls(user_id, origin_url, short_url) VALUES($1,$2,$3)", userID, longURL, shortURL)
	if err != nil {
		//return fmt.Errorf(`%w`, err)
		return err
	}
	return nil
}

func (store StoreDB) SaveBatchLink(ctx context.Context, batch []ElemBatch, userID string) error {
	if store.db == nil {
		return ErrDBConnection //errors.New("you haven`t opened the database connection")
	}
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.StmtContext(ctx, store.insertStmt)
	defer stmt.Close()

	for _, v := range batch {
		longURL, _ := store.GetByID(ctx, v.ShortURL)
		if longURL != "" {
			continue
		}
		if _, err = stmt.Exec(userID, v.OriginURL, v.ShortURL); err != nil {
			if errRB := tx.Rollback(); errRB != nil {
				return fmt.Errorf("update drivers: unable to rollback: %w", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update drivers: unable to commit: %w", err)
	}
	return nil
}

func (store StoreDB) GetByID(ctx context.Context, id string) (string, error) {
	if store.db == nil {
		return "", ErrDBConnection //errors.New("you haven`t opened the database connection")
	}
	var row RowTabURL
	err := store.db.GetContext(ctx, &row, "SELECT * FROM urls WHERE short_url = $1 LIMIT 1", id)
	if err != nil { //sql.ErrNoRows
		return "", err
	}
	if row.Deleted {
		return "", ErrURLDeleted
	}
	return row.OriginalURL, nil
}

func (store *StoreDB) NewUserID(ctx context.Context) string {
	userID := generateUserID()

	result, err := store.db.ExecContext(ctx, "INSERT INTO users_token(user_id) VALUES($1)", userID)
	if err != nil {
		return ""
	}

	id, err := result.LastInsertId()
	if err == nil {
		fmt.Printf("Идентификатор новой записи usersToken %d", id)
	}
	return userID
}

func (store StoreDB) UserIDIsExist(ctx context.Context, UserID string) bool {
	rows, err := store.db.QueryContext(ctx, "SELECT user_id FROM users_token WHERE user_id = $1", UserID)
	if err != nil || rows.Err() != nil {
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func (store StoreDB) GetUserURLs(ctx context.Context, UserID string) ([]PairURL, error) {
	var urls []PairURL
	err := store.db.SelectContext(ctx, &urls, "SELECT origin_url, short_url FROM urls WHERE user_id = $1", UserID)
	if err != nil {
		return nil, fmt.Errorf("Error exec query: " + err.Error())
	}
	return urls, nil
}

func (store StoreDB) GetUserMapURLs(ctx context.Context, UserID string) (map[string]string, error) {
	urls := make(map[string]string)
	originURL := ""
	shortURL := ""
	rows, err := store.db.QueryContext(ctx, "SELECT origin_url, short_url FROM urls WHERE user_id = $1", UserID)
	if err != nil || rows.Err() != nil {
		return nil, fmt.Errorf("Error exec query: " + err.Error())
	}
	for rows.Next() {
		err := rows.Scan(&originURL, &shortURL)
		if err != nil {
			return nil, fmt.Errorf("Error scan row: " + err.Error())
		}
		urls[shortURL] = originURL
	}
	return urls, nil
}

func (store StoreDB) DeleteURLs(ctx context.Context, masID []string) error {
	if store.db == nil {
		return ErrDBConnection
	}
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}
	stmt := tx.StmtContext(ctx, store.makeDelStmt)
	defer stmt.Close()

	for _, val := range masID {
		if _, err = stmt.Exec(val); err != nil {
			if errRB := tx.Rollback(); errRB != nil {
				return fmt.Errorf("update drivers: unable to rollback: %w", err)
			}
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update drivers: unable to commit: %w", err)
	}
	return nil
}
