package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strconv"
	//"url-shortener/pkg/handler"
)

type Storager interface {
	SaveLink(ctx context.Context, shortLink, longLink, userID string) error
	SaveBatchLink(ctx context.Context, batch []ElemBatch, userID string) error
	GetByID(ctx context.Context, id string) (string, error)
	GetAll(ctx context.Context) map[string]UserURL
	NewUserID(ctx context.Context) string
	UserIdIsExist(ctx context.Context, userID string) bool //Проверка что такой User Id выдавался
	GetUserURLs(ctx context.Context, userID string) ([]PairURL, error)
	GetUserMapURLs(ctx context.Context, UserID string) (map[string]string, error)
	DeleteURLs(ctx context.Context, masID []string) error
	Ping() bool
	Close()
}

type StoreConfig struct {
	BaseURL         string
	DBDNS           string
	FilestoragePath string
}

type PairURL struct {
	ShortURL    string `json:"short_url,omitempty" db:"short_url"`
	OriginalURL string `json:"original_url,omitempty" db:"origin_url"`
}

type UserURL struct {
	OriginURL string `json:"originUrl"`
	UserID    string `json:"userId"`
	Deleted   bool   `json:"deleted"`
}

type ElemBatch struct {
	CoreID    string `json:"correlation_id"`
	OriginURL string `json:"original_url"`
	ShortURL  string `json:"short_url"`
}

type UserArrayURL struct {
	UserID   string
	ArrayURL []string
}

func ConfigurateStorage(c *StoreConfig) Storager {
	postgreDB, err := NewStoreDB(c)
	if err != nil || !postgreDB.Ping() {
		memoryDB := NewMemoryRep(c.FilestoragePath, c.BaseURL)
		return memoryDB
	}
	return postgreDB
}

//Функция которая принимает в качестве аргумента именно интерфейс
func AddToCollection(ctx context.Context, m Storager, longURL, userID string) (s string, err error) {
	shortURL := generateIdentify(longURL)
	return shortURL, m.SaveLink(ctx, shortURL, longURL, userID)
}

func AddToCollectionBatch(ctx context.Context, m Storager, batch []ElemBatch, userID string) error {
	for ind, el := range batch {
		batch[ind].ShortURL = generateIdentify(el.OriginURL)
	}
	return m.SaveBatchLink(ctx, batch, userID)
}

func generateIdentify(s string) string {
	h := crc32.NewIEEE()
	h.Write([]byte(s))
	hashSum := h.Sum32()
	sHash := strconv.Itoa(int(hashSum))
	masRune := []rune(sHash)
	for i, value := range masRune {
		//masRune[i] = rune(int(value) + 49 + rand.Intn(15))
		masRune[i] = value + 49
	}
	res := string(masRune)
	return res
}

func generateUserID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Printf("Не удалось сформировать идентификатор пользователя: %v\n", err)
		return ""
	}
	id := hex.EncodeToString(b)
	return id
}
