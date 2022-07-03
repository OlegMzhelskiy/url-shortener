package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type MemoryRep struct {
	db              map[string]UserURL
	fileStoragePath string
	usersId         map[string]int
	baseUrl         string
}

func NewMemoryRep(fileStoragePath, baseUrl string) *MemoryRep {
	rep := &MemoryRep{
		db:              make(map[string]UserURL),
		fileStoragePath: fileStoragePath,
		usersId:         make(map[string]int),
		baseUrl:         baseUrl,
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

func (m MemoryRep) GetAll() map[string]UserURL {
	return m.db
}

func (m MemoryRep) SaveLink(shortURL, longURL, userId string) error {
	_, ok := m.db[shortURL]
	if !ok {
		m.db[shortURL] = UserURL{longURL, userId}
		err := m.WriteRepoFromFile() //При сохранении нового URL запишем в файл
		if err != nil {
			return errors.New(fmt.Sprintf("Ошибка записи в файл: %s", err))
		}
	}
	return nil
}

func (m MemoryRep) SaveBatchLink(batch []ElemBatch, userId string) error {
	for _, el := range batch {
		err := m.SaveLink(el.ShortUrl, el.OriginUrl, userId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m MemoryRep) GetByID(id string) (string, error) {
	val, ok := m.db[id]
	if !ok {
		return "", errors.New("not found")
	}
	return val.OriginUrl, nil
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
		return err //panic(err)
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

func (m *MemoryRep) NewUserID() string {
	id := generateUserID()
	m.usersId[id]++
	return id
	//return base64.StdEncoding.EncodeToString(b)
}

func (m *MemoryRep) UserIdIsExist(UserId string) bool {
	return m.usersId[UserId] > 0
}

func (m MemoryRep) GetUserUrls(UserId string) []PairURL {
	masUrls := make([]PairURL, 0)
	for key, val := range m.db {
		if val.UserId == UserId {
			newPair := PairURL{m.baseUrl + key, val.OriginUrl}
			masUrls = append(masUrls, newPair)
		}
	}
	return masUrls
}

func (m MemoryRep) Ping() bool {
	return false
}

func (m MemoryRep) Close() {

}
