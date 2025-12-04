package fileobs

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Skifskii/link-shortener/internal/service/audit"
)

type FileObserver struct {
	filePath string
	mu       sync.Mutex
}

func New(filePath string) *FileObserver {
	return &FileObserver{
		filePath: filePath,
	}
}

func (f *FileObserver) Update(e *audit.Event) {
	if e == nil {
		return
	}
	f.addEventToFile(*e)
}

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
