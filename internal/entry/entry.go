package entry

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tsubasaogawa/hatenablog-flickr-to-s3-converter/internal/url"
)

var (
	rFlickrATag      = regexp.MustCompile(`<a.*href="https?://www\.flickr\.com/(?:photos/\w+/\d+/in/[^"]+|gp/\w+/\w+)"[^>]*>`)
	rFlickrScriptTag = regexp.MustCompile(`<script.*src="//embedr.flickr.com/assets/client-code.js"[^>]*></script>`)
	rEntryPathSuffix = regexp.MustCompile(`entry[/\\]\d{4}[/\\]\d{2}[/\\]\d{2}[/\\]\d{6}.md$`)
)

type Entry struct {
	File, Body, BackupFile string
}

func NewEntry(file, backupDir string, dryrun bool) (Entry, error) {
	textb, err := os.ReadFile(file)
	if err != nil {
		return Entry{}, err
	}

	backupFile := ""
	if backupDir != "" && !dryrun {
		backupFile, err = backup(file, backupDir, &textb)
		if err != nil {
			return Entry{}, err
		}
	}

	return Entry{
		File:       file,
		Body:       string(textb),
		BackupFile: backupFile,
	}, nil
}

func (entry *Entry) Replace(replaceUrlPairs url.Urls) {
	entry.Body = strings.NewReplacer(replaceUrlPairs.Flatten()...).Replace(
		rFlickrScriptTag.ReplaceAllString(
			rFlickrATag.ReplaceAllString(entry.Body, `<a tabindex="-1">`),
			"",
		),
	)
}

func (entry *Entry) Save() error {
	return os.WriteFile(entry.File, []byte(entry.Body), os.ModePerm)
}

func backup(fromFile, toDirBase string, data *[]byte) (string, error) {
	entryPathSuffix := rEntryPathSuffix.FindString(fromFile)
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
	return backupFile, os.WriteFile(backupFile, *data, os.ModePerm)
}
