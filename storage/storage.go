package storage

import (
	"errors"
	"hash/crc32"
	"strconv"
)

type Storager interface {
	SaveLink(shortLink, longLink string) error
	GetByID(string) (string, error)
}

type MemoryRep struct {
	db map[string]string
}

func NewMemoryRep() *MemoryRep {
	return &MemoryRep{
		db: make(map[string]string),
	}
}

func (m MemoryRep) SaveLink(shortURL, longURL string) error {
	_, ok := m.db[shortURL]
	if !ok {
		m.db[shortURL] = longURL
	}
	return nil
}

func (m MemoryRep) GetByID(id string) (string, error) {
	val, ok := m.db[id]
	if !ok {
		return "", errors.New("not found")
	}
	return val, nil
}

//Функция которая принимает в качестве аргумента именно интерфейс
func AddToCollection(rep Storager, longURL string) (s string, err error) {
	shortURL := generateIdentify(longURL)
	return shortURL, rep.SaveLink(shortURL, longURL)
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
