package config_test

import (
	"os"
	"testing"

	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/config"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestNewConfig(t *testing.T) {
	tests := map[string]struct {
		tomlFile string
		hasErr   bool
	}{
		"Not found toml file": {"hoge/fuga", true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := config.NewConfig(tt.tomlFile); tt.hasErr == (err == nil) {
				t.Errorf("got: %#v, want: %#v\n", err != nil, tt.hasErr)
			}
		})
	}
}
