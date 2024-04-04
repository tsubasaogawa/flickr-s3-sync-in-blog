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
		if f, err := os.Stat(backupDir); os.IsNotExist(err) || !f.IsDir() {
			if err = os.MkdirAll(backupDir, os.ModePerm); err != nil {
				return Entry{}, err
			}
		}
		backupFile := filepath.Join(backupDir, filepath.Base(file)) + ".bak"
		if err = os.WriteFile(backupFile, textb, os.ModePerm); err != nil {
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
