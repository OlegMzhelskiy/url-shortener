package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	//"database/sql"
	"fmt"
	_ "github.com/jackc/pgx"
	//_ "github.com/lib/pq"
)

type StoreDB struct {
	db     *sqlx.DB
	config *StoreConfig
	//dbdns string
}

type StoreConfig struct {
	BaseUrl string
	DBDNS   string
}

func NewStoreDB(config *StoreConfig) (*StoreDB, error) {
	db, err := sqlx.Open("postgres", config.DBDNS)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls(
		originUrl TEXT UNIQUE NOT NULL,
		shortUrl TEXT NOT NULL,
		userId TEXT NOT NULL);
		CREATE TABLE IF NOT EXISTS usersToken(
		userId TEXT UNIQUE NOT NULL);`)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &StoreDB{db, config}, nil
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

	rows, err := store.db.QueryContext(ctx, "SELECT userId, originUrl, shortUrl FROM urls")

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
	url, _ := store.GetByID(shortURL)

	//Если записи с таким url нет то добавим
	if url == "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := store.db.ExecContext(ctx, "INSERT INTO urls(userId, originUrl, shortUrl) VALUES($1,$2,$3)", userId, longURL, shortURL)
		return err
	}
	return nil
}

func (store StoreDB) GetByID(id string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var originUrl string
	err := store.db.QueryRowContext(ctx, "SELECT originUrl FROM urls WHERE shortUrl = $1 LIMIT 1", id).Scan(&originUrl)
	if err != nil { //sql.ErrNoRows
		return "", err
	}
	return originUrl, nil
}

func (store *StoreDB) NewUserID() string {
	userId := generateUserID()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := store.db.ExecContext(ctx, "INSERT INTO usersToken(userId) VALUES($1)", userId)
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

	rows, err := store.db.QueryContext(ctx, "SELECT userId FROM usersToken WHERE userId = $1", UserId)
	if err != nil {
		return false
	}
	defer rows.Close()
	return rows.Next()
}

func (store StoreDB) GetUserUrls(UserId string) []PairURL {
	var urls = []PairURL{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := store.db.QueryContext(ctx, "SELECT originUrl, shortUrl FROM urls WHERE userId = $1", UserId)

	if err == nil {
		defer rows.Close()

		for rows.Next() {
			var (
				origin string
				short  string
			)
			err = rows.Scan(&origin, &short)
			if err != nil {
				return urls
			}
			newPair := PairURL{store.config.BaseUrl + origin, short}
			urls = append(urls, newPair)
		}
		err = rows.Err()
		if err != nil {
			return urls
		}
	}
	return urls
}
