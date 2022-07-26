package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type MemoryRep struct {
	db              map[string]UserURL
	fileStoragePath string
	usersID         map[string]int
	baseURL         string
	file            *os.File
}

func (m MemoryRep) Close() {
	if m.file == nil {
		m.file.Close()
	}
}

func NewMemoryRep(fileStoragePath, baseURL string) *MemoryRep {
	var errOpen error
	var file *os.File

	if len(fileStoragePath) > 0 {
		file, errOpen = os.OpenFile(fileStoragePath, os.O_RDWR|os.O_CREATE, 0777)
	}

	rep := &MemoryRep{
		db:              make(map[string]UserURL),
		fileStoragePath: fileStoragePath,
		usersID:         make(map[string]int),
		baseURL:         baseURL,
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

func (m MemoryRep) GetAll(ctx context.Context) map[string]UserURL {
	return m.db
}

func (m MemoryRep) SaveLink(ctx context.Context, shortURL, longURL, userID string) error {
	_, ok := m.db[shortURL]
	if !ok {
		usURL := UserURL{longURL, userID, false}
		m.db[shortURL] = usURL

		//err := m.WriteRepoFromFile() //При сохранении нового URL запишем в файл
		elem := elementCollection{shortURL, usURL}
		err := m.WriteElementFromFile(elem)

		if err != nil {
			return fmt.Errorf("error write a file: %s", err)
		}
	}
	return nil
}

func (m MemoryRep) SaveBatchLink(ctx context.Context, batch []ElemBatch, userID string) error {
	for _, el := range batch {
		err := m.SaveLink(ctx, el.ShortURL, el.OriginURL, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m MemoryRep) GetByID(ctx context.Context, id string) (string, error) {
	val, ok := m.db[id]
	if !ok {
		return "", errors.New("not found")
	}
	return val.OriginURL, nil
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
	data, err := json.Marshal(m.db)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	writer := bufio.NewWriter(m.file)
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("error write a file: %s", err)
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

func (m *MemoryRep) NewUserID(ctx context.Context) string {
	id := generateUserID()
	m.usersID[id]++
	return id
	//return base64.StdEncoding.EncodeToString(b) //
}

func (m *MemoryRep) UserIDIsExist(ctx context.Context, userID string) bool {
	return m.usersID[userID] > 0
}

func (m MemoryRep) GetUserURLs(ctx context.Context, userID string) ([]PairURL, error) {
	masUrls := make([]PairURL, 0)
	for key, val := range m.db {
		if val.UserID == userID {
			newPair := PairURL{ShortURL: m.baseURL + key, OriginalURL: val.OriginURL}
			masUrls = append(masUrls, newPair)
		}
	}
	return masUrls, nil
}

func (m MemoryRep) GetUserMapURLs(ctx context.Context, userID string) (map[string]string, error) {
	urls := make(map[string]string)
	for key, val := range m.db {
		if val.UserID == userID {
			urls[key] = val.OriginURL
		}
	}
	return urls, nil
}

func (m MemoryRep) DeleteURLs(ctx context.Context, masID []string) error {
	for _, id := range masID {
		el := m.db[id]
		if !el.Deleted {
			m.db[id] = UserURL{el.OriginURL, el.UserID, true}
		}
	}
	return nil
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
