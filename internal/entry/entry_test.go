package entry_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/config"
	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/entry"
)

var (
	TEST_PATH string
)

func init() {
	_, TEST_PATH, _, _ = runtime.Caller(0)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestNewEntry(t *testing.T) {
	tests := map[string]struct {
		file   string
		hasErr bool
	}{
		"File not found": {"hoge/fuga", true},
		"Existing file":  {TEST_PATH, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := entry.NewEntry(tt.file, &config.Config{}); tt.hasErr == (err == nil) {
				t.Errorf("got: %#v, want: %#v\n", err != nil, tt.hasErr)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	t.Skipf("TODO\n")
}

func TestBackup(t *testing.T) {
	t.Skipf("TODO\n")
}
