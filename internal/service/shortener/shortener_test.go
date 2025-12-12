package shortener

import (
	"testing"

	"github.com/Skifskii/link-shortener/internal/repository/inmemory"
)

func BenchmarkGenerateShortCode(b *testing.B) {
	repo := inmemory.New()
	s := New("http://localhost:8080", 8, repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.generateShortCode(); err != nil {
			b.Fatal(err)
		}
	}
}
