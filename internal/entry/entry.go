package entry

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tsubasaogawa/hatenablog-flickr-to-s3-converter/internal/flickr"
	"github.com/tsubasaogawa/hatenablog-flickr-to-s3-converter/internal/url"
)

var (
	reEntryPathSuffix = regexp.MustCompile(`entry[/\\]\d{4}[/\\]\d{2}[/\\]\d{2}[/\\]\d{6}.md$`)
)

type Entry struct {
	File, body, NewBody string
}

func NewEntry(file, backupDir string, dryrun bool) (Entry, error) {
	textb, err := os.ReadFile(file)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		File: file,
		body: string(textb),
	}, nil
}

func (entry *Entry) FindFlickrUrls() []string {
	return flickr.ReUrl.FindAllString(entry.body, -1)
}

func (entry *Entry) Replace(replaceUrlPairs url.Urls) {
	entry.NewBody = strings.NewReplacer(replaceUrlPairs.Flatten()...).Replace(
		flickr.ReScriptTag.ReplaceAllString(
			flickr.ReATag.ReplaceAllString(entry.body, `<a tabindex="-1">`),
			"",
		),
	)
}

func (entry *Entry) Save() error {
	return os.WriteFile(entry.File, []byte(entry.NewBody), os.ModePerm)
}

func (entry *Entry) Backup(fromFile, toDirBase string) (string, error) {
	entryPathSuffix := reEntryPathSuffix.FindString(fromFile)
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
	return backupFile, os.WriteFile(backupFile, []byte(entry.body), os.ModePerm)
}
