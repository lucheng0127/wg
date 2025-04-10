package core

import (
	"fmt"
	"testing"
)

func TestNewRandomKey(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "generate",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRandomKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRandomKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Printf("%+v", got)
		})
	}
}
