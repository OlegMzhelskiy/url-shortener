package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
	//"database/sql"
	"fmt"
	_ "github.com/jackc/pgx"
	//_ "github.com/lib/pq"
)

type StoreDB struct {
	db         *sqlx.DB
	config     *StoreConfig
	insertStmt *sql.Stmt
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
		user_id TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS users_token(
		user_id TEXT UNIQUE NOT NULL);`)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(time.Second * 30)
	db.SetConnMaxLifetime(time.Minute * 2)

	insertStmt, _ := db.Prepare("INSERT INTO urls(user_id, origin_url, short_url) VALUES($1,$2,$3)")

	return &StoreDB{db, config, insertStmt}, nil
}

func (store *StoreDB) Close() {
	store.db.Close()
}

func (store *StoreDB) Ping() bool {
	if err := store.db.Ping(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (store StoreDB) GetAll() map[string]UserURL {
	urls := make(map[string]UserURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := store.db.QueryContext(ctx, "SELECT user_id, origin_url, short_url FROM urls")

	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var (
				us string
				or string
				sh string
			)
			err = rows.Scan(&us, &or, &sh)
			if err != nil {
				return urls
			}
			urls[sh] = UserURL{or, us}
		}
		err = rows.Err()
		if err != nil {
			return urls
		}
	}
	return urls
}

func (store StoreDB) SaveLink(shortURL, longURL, userId string) error {
	if store.db.DB == nil {
		return errors.New("You haven`t opened the database connection")
	}
	url, _ := store.GetByID(shortURL)

	//Если записи с таким url нет то добавим
	if url == "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := store.db.ExecContext(ctx, "INSERT INTO urls(user_id, origin_url, short_url) VALUES($1,$2,$3)", userId, longURL, shortURL)
		return err
	}
	return nil
}

func (store StoreDB) SaveBatchLink(batch []ElemBatch, userId string) error {
	if store.db.DB == nil {
		return errors.New("You haven`t opened the database connection")
	}
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	//stmt := tx.StmtContext(ctx, insertStmt)
	stmt, err := tx.Prepare("INSERT INTO urls(user_id, origin_url, short_url) VALUES($1,$2,$3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, v := range batch {
		longUrl, _ := store.GetByID(v.ShortUrl)
		if longUrl != "" {
			continue
		}
		if _, err = stmt.Exec(userId, v.OriginUrl, v.ShortUrl); err != nil {
			if err = tx.Rollback(); err != nil {
				log.Fatalf("update drivers: unable to rollback: %v", err)
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("update drivers: unable to commit: %v", err)
		return err
	}
	return nil
}

func (store StoreDB) GetByID(id string) (string, error) {
	if store.db.DB == nil {
		return "", errors.New("You haven`t opened the database connection")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var originUrl string
	//err := store.db.QueryRowContext(ctx, "SELECT origin_url FROM urls WHERE short_url = $1 LIMIT 1", id).Scan(&originUrl)
	err := store.db.GetContext(ctx, &originUrl, "SELECT origin_url FROM urls WHERE short_url = $1 LIMIT 1", id)

	if err != nil { //sql.ErrNoRows
		return "", err
	}
	return originUrl, nil
}

func (store *StoreDB) NewUserID() string {
	userId := generateUserID()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := store.db.ExecContext(ctx, "INSERT INTO users_token(user_id) VALUES($1)", userId)
	if err != nil {
		return ""
	}

	id, err := result.LastInsertId()
	if err == nil {
		fmt.Println("Идентификатор новой записи usersToken %s", id)
	}
	return userId
}

func (store StoreDB) UserIdIsExist(UserId string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := store.db.QueryContext(ctx, "SELECT user_id FROM users_token WHERE user_id = $1", UserId)
	if err != nil {
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func (store StoreDB) GetUserUrls(UserId string) []PairURL {
	var urls []PairURL

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//rows, err := store.db.QueryContext(ctx, "SELECT origin_url, short_url FROM urls WHERE userId = $1", UserId)
	//if err == nil {
	//	defer rows.Close()
	//
	//	for rows.Next() {
	//
	//		var (
	//			origin string
	//			short  string
	//		)
	//		err = rows.Scan(&origin, &short)
	//		if err != nil {
	//			return urls
	//		}
	//		newPair := PairURL{store.config.BaseUrl + origin, short}
	//		urls = append(urls, newPair)
	//	}
	//	err = rows.Err()
	//	if err != nil {
	//		return urls
	//	}
	//}

	err := store.db.SelectContext(ctx, &urls, "SELECT origin_url, short_url FROM urls WHERE user_id = $1", UserId)
	if err != nil {
		fmt.Println("Error exec query: " + err.Error())
	}
	return urls
}
