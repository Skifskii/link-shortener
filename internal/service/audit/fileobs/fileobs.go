// Package fileobs реализует наблюдатель аудита, который сохраняет события в файл.
package fileobs

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Skifskii/link-shortener/internal/service/audit"
)

// FileObserver записывает события аудита в указанный файл.
type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

// New создаёт FileObserver для записи событий в файл filePath.
func New(filePath string) *FileObserver {
	return &FileObserver{
		filePath: filePath,
	}
}

// Update получает событие и сохраняет его в файл (в виде JSON-строки).
func (f *FileObserver) Update(e *audit.Event) {
	if e == nil {
		return
	}
	f.addEventToFile(*e)
}

// addEventToFile выполняет потокобезопасную запись события в файл.
func (f *FileObserver) addEventToFile(e audit.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}
