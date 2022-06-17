package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"strconv"
)

type Storager interface {
	SaveLink(shortLink, longLink string) error
	GetByID(string) (string, error)
	GetAll() map[string]string
}

type MemoryRep struct {
	db              map[string]string
	fileStoragePath string
}

//Инициализация
func NewMemoryRep(fileStoragePath string) *MemoryRep {
	rep := &MemoryRep{
		db:              make(map[string]string),
		fileStoragePath: fileStoragePath,
	}
	//Если заполнен путь к файлу то читаем сохраненные URL
	if len(fileStoragePath) > 0 {
		err := rep.ReadRepoFromFile()
		if err != nil {
			fmt.Printf("Ошибка чтения файла %s \n", err)
		}
	}
	return rep
}

func (m MemoryRep) GetAll() map[string]string {
	return m.db
}

func (m MemoryRep) SaveLink(shortURL, longURL string) error {
	_, ok := m.db[shortURL]
	if !ok {
		m.db[shortURL] = longURL
		err := m.WriteRepoFromFile() //При сохранении нового URL запишем в файл
		if err != nil {
			return errors.New(fmt.Sprintf("Ошибка записи в файл: %s", err))
		}
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
func AddToCollection(m Storager, longURL string) (s string, err error) {
	shortURL := generateIdentify(longURL)
	return shortURL, m.SaveLink(shortURL, longURL)
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

//Сохраняем Rep в файл
func (m MemoryRep) WriteRepoFromFile() error {
	if len(m.fileStoragePath) == 0 {
		return nil
	}
	//каждый раз перезаписываем файл
	file, err := os.OpenFile(m.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	data, err := json.Marshal(m.db)
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if _, err := writer.Write(data); err != nil {
		return err
	}
	// записываем буфер в файл
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

//Читаем данные из файла
func (m *MemoryRep) ReadRepoFromFile() error {
	if len(m.fileStoragePath) == 0 {
		return nil
	}
	file, err := os.OpenFile(m.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	//reader := bufio.NewReader(file)
	//data, err := reader.ReadBytes()
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	data := scanner.Bytes()

	if data != nil {
		err = json.Unmarshal(data, &m.db)
		if err != nil {
			return err
		}
	}
	return nil
}
