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
	File, Body string
}

func NewEntry(file, backupDir string, dryrun bool) (Entry, error) {
	textb, err := os.ReadFile(file)
	if err != nil {
		return Entry{}, err
	}

	if backupDir != "" && !dryrun {
		if err := backup(file, backupDir, textb); err != nil {
			return Entry{}, err
		}
	}

	return Entry{
		File: file,
		Body: string(textb),
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

// TODO: pointer argument `data`
func backup(fromFile, toDirBase string, data []byte) error {
	entryPathSuffix := rEntryPathSuffix.FindString(fromFile)
	if entryPathSuffix == "" {
		entryPathSuffix = filepath.Base(fromFile)
	}
	toFile := filepath.Join(toDirBase, entryPathSuffix)
	toDir := filepath.Dir(toFile)

	if f, err := os.Stat(toDir); os.IsNotExist(err) || !f.IsDir() {
		if err = os.MkdirAll(toDir, os.ModePerm); err != nil {
			return err
		}
	}

	return os.WriteFile(toFile+".bak", data, os.ModePerm)
}
