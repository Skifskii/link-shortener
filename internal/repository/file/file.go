package file

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

var errEmptyFilepath = errors.New("filepath is empty")

type LinkRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileRepo struct {
	*fileWriter
	*fileReader
	mu sync.Mutex
}

type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

func newFileWriter(filepath string) (*fileWriter, error) {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &fileWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (fw *fileWriter) writeLinkRecord(lr *LinkRecord) error {
	data, err := json.Marshal(lr)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := fw.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := fw.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return fw.writer.Flush()
}

type fileReader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func newFileReader(filepath string) (*fileReader, error) {
	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &fileReader{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (fr *fileReader) readLinkRecordByShortURL(shortURL string) (*LinkRecord, error) {
	for fr.scanner.Scan() {
		data := fr.scanner.Bytes()

		var lr LinkRecord
		if err := json.Unmarshal(data, &lr); err != nil {
			return nil, err
		}

		if lr.ShortURL == shortURL {
			return &lr, nil
		}
	}

	if err := fr.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("LinkRecord with short_url %q not found", shortURL)
}

func (fr *fileReader) countRecords() (int, error) {
	if _, err := fr.file.Seek(0, 0); err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(fr.file)
	count := 0
	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	if _, err := fr.file.Seek(0, 0); err != nil {
		return 0, err
	}
	fr.scanner = bufio.NewScanner(fr.file)

	return count, nil
}

func NewFileRepo(filepath string) (*FileRepo, error) {
	if filepath == "" {
		return nil, errEmptyFilepath
	}

	fileRepo := FileRepo{}
	var err error

	fileRepo.fileWriter, err = newFileWriter(filepath)
	if err != nil {
		return nil, err
	}

	fileRepo.fileReader, err = newFileReader(filepath)
	if err != nil {
		return nil, err
	}

	return &fileRepo, nil
}

func (fr *FileRepo) Save(short, original string) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	lenFile, err := fr.countRecords()
	if err != nil {
		return err
	}

	return fr.writeLinkRecord(&LinkRecord{
		UUID:        fmt.Sprint(lenFile),
		ShortURL:    short,
		OriginalURL: original,
	})
}

func (fr *FileRepo) Get(short string) (string, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	lr, err := fr.readLinkRecordByShortURL(short)
	if err != nil {
		return "", err
	}

	return lr.OriginalURL, nil
}
