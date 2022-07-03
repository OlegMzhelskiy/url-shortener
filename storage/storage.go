package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strconv"
	//"url-shortener/pkg/handler"
)

type Storager interface {
	SaveLink(shortLink, longLink, userId string) error
	SaveBatchLink(batch []ElemBatch, userId string) error
	GetByID(string) (string, error)
	GetAll() map[string]UserURL
	NewUserID() string
	UserIdIsExist(userId string) bool //Проверка что такой User Id выдавался
	GetUserUrls(userId string) []PairURL
	Ping() bool
}

type PairURL struct {
	ShortUrl    string `json:"short_url,omitempty" db:"short_url"`
	OriginalUrl string `json:"original_url,omitempty" db:"origin_url"`
}

type UserURL struct {
	OriginUrl string `json:"originUrl"`
	UserId    string `json:"userId"`
}

type ElemBatch struct {
	CoreId    string `json:"correlation_id"`
	OriginUrl string `json:"original_url"`
	ShortUrl  string `json:"short_url"`
}

//Функция которая принимает в качестве аргумента именно интерфейс
func AddToCollection(m Storager, longURL, userId string) (s string, err error) {
	shortURL := generateIdentify(longURL)
	return shortURL, m.SaveLink(shortURL, longURL, userId)
}

func AddToCollectionBatch(m Storager, batch []ElemBatch, userId string) error {
	for ind, el := range batch {
		batch[ind].ShortUrl = generateIdentify(el.OriginUrl)
	}
	return m.SaveBatchLink(batch, userId)
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
