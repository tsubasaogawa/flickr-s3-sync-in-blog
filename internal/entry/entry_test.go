package entry_test

import (
	"os"
	"path/filepath"
	"regexp"
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
	tmpdir, err := os.MkdirTemp(os.TempDir(), "entries")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	f, err := os.CreateTemp(tmpdir, "entry")
	if err != nil {
		t.Fatal(err)
	}
	fileSuffixPtn := filepath.Join(`entries.+`, `entry.+`) + `$`
	reFileSuffix := regexp.MustCompile(fileSuffixPtn)
	fileSuffix := reFileSuffix.FindString(f.Name())
	if fileSuffix == "" {
		t.Errorf("Cannot match a temp file name with the suffix pattern")
	}
	want := filepath.Join(os.TempDir(), fileSuffix) + ".bak"

	c := config.Config{}
	c.Regex.EntryPath = map[string]string{}
	c.Regex.EntryPath["suffix"] = fileSuffixPtn

	e, err := entry.NewEntry(f.Name(), &c)
	if err != nil {
		t.Fatal(err)
	}

	got, err := e.Backup(f.Name(), os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got: %#v, want: %#v\n", got, want)
	}
}
