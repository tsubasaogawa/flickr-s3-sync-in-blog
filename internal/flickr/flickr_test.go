package flickr_test

import (
	"os"
	"testing"

	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/config"
	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/flickr"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestNewFlickr(t *testing.T) {
	tests := map[string]struct {
		url  string
		want *flickr.Flickr
	}{
		"Given url is set": {"https://foo/bar", &flickr.Flickr{Url: "https://foo/bar"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := flickr.NewFlickr(tt.url, &config.Config{})
			if f.Url != tt.want.Url {
				t.Errorf("got: %#v, want: %#v\n", f.Url, tt.want.Url)
			}
		})
	}
}
