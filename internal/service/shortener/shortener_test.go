package shortener

import "testing"

func TestGenerateShortURL(t *testing.T) {
	tests := []struct {
		name    string
		wantLen int
	}{
		{
			name:    "common case",
			wantLen: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLen := len(GenerateShortURL()); gotLen != tt.wantLen {
				t.Errorf("GenerateShortURL() = %v, want %v", gotLen, tt.wantLen)
			}
		})
	}
}
