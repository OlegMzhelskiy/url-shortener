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
	file            *os.File
}

func (m MemoryRep) Close() {
	if m.file == nil {
		m.file.Close()
	}
}

func NewMemoryRep(fileStoragePath, baseUrl string) *MemoryRep {
	var errOpen error
	var file *os.File

	if len(fileStoragePath) > 0 {
		file, errOpen = os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE, 0777)
	}

	rep := &MemoryRep{
		db:              make(map[string]UserURL),
		fileStoragePath: fileStoragePath,
		usersId:         make(map[string]int),
		baseUrl:         baseUrl,
		file:            file,
	}

	if errOpen != nil {
		fmt.Printf("File opening error %s \n", errOpen)
	} else {
		err := rep.ReadRepoFromFile()
		if err != nil {
			fmt.Printf("File reading error  %s \n", err)
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
		usUrl := UserURL{longURL, userId}
		m.db[shortURL] = usUrl

		//err := m.WriteRepoFromFile() //При сохранении нового URL запишем в файл
		elem := elementCollection{shortURL, usUrl}
		err := m.WriteElementFromFile(elem)

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

func (m MemoryRep) WriteElementFromFile(elem elementCollection) error {
	if m.file == nil {
		return nil
	}

	data, err := json.Marshal(elem)
	if err != nil {
		return err
	}

	m.file.Seek(0, 0)

	scanner := bufio.NewScanner(m.file)
	scanner.Scan()
	dataScan := scanner.Bytes()

	//var dataFile []byte
	//reader := bufio.NewReader(m.file)
	//size, err := reader.Read(dataFile)
	//if err != nil {
	//	return err
	//}
	//
	//size, _ = m.file.Read(dataFile)

	size := len(dataScan)

	var off int64 = 0
	if size != 0 {
		ndata := data[1 : len(data)-1]           //убираем {}
		data = []byte("," + string(ndata) + "}") //пишем в конец
		off = int64(size) - 1
	}

	if _, err := m.file.WriteAt(data, off); err != nil {
		return err
	}
	return nil
}

//Сохраняем Rep в файл
func (m MemoryRep) WriteRepoFromFile() error {
	//if len(m.fileStoragePath) == 0 {
	if m.file == nil {
		return nil
	}
	//каждый раз перезаписываем файл
	//file, err := os.OpenFile(m.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	//if err != nil {
	//	return err
	//}
	//defer file.Close()

	data, err := json.Marshal(m.db)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	writer := bufio.NewWriter(m.file)
	if _, err := writer.Write(data); err != nil {
	}
	// записываем буфер в файл
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

//Читаем данные из файла
func (m *MemoryRep) ReadRepoFromFile() error {
	if m.file == nil {
		return nil
	}
	//file, err := os.OpenFile(m.fileStoragePath, os.O_RDONLY|os.O_CREATE, 0777)
	//if err != nil {
	//	return err
	//}
	//defer file.Close()

	scanner := bufio.NewScanner(m.file)
	scanner.Scan()
	data := scanner.Bytes()

	if data != nil {
		err := json.Unmarshal(data, &m.db)
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

type elementCollection struct {
	short   string
	UserURL UserURL
}

func (ec elementCollection) MarshalJSON() ([]byte, error) {
	url, err := json.Marshal(ec.UserURL)
	if err != nil {
		return []byte(""), err
	}
	res := "{\"" + ec.short + "\":" + string(url) + "}"
	return []byte(res), nil
}
