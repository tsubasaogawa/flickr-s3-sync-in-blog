package entry

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/config"
	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/flickr"
	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/url"
)

type Entry struct {
	file, Body, NewBody string
	config              *config.Config
}

func NewEntry(file string, conf *config.Config) (Entry, error) {
	textb, err := os.ReadFile(file)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		file:   file,
		Body:   string(textb),
		config: conf,
	}, nil
}

func (entry *Entry) Replace(replaceUrlPairs url.Urls) {
	entry.NewBody = strings.NewReplacer(replaceUrlPairs.Flatten()...).Replace(
		flickr.ReplaceNeedlessTags(entry.Body, entry.config),
	)
}

func (entry *Entry) Save() error {
	return os.WriteFile(entry.file, []byte(entry.NewBody), os.ModePerm)
}

func (entry *Entry) Backup(fromFile, toDirBase string) (string, error) {
	entryPathSuffix := regexp.MustCompile(entry.config.Regex.EntryPath["suffix"]).FindString(fromFile)
	if entryPathSuffix == "" {
		entryPathSuffix = filepath.Base(fromFile)
	}
	toFile := filepath.Join(toDirBase, entryPathSuffix)
	toDir := filepath.Dir(toFile)

	if f, err := os.Stat(toDir); os.IsNotExist(err) || !f.IsDir() {
		if err = os.MkdirAll(toDir, os.ModePerm); err != nil {
			return "", err
		}
	}

	backupFile := toFile + ".bak"
	return backupFile, os.WriteFile(backupFile, []byte(entry.Body), os.ModePerm)
}
