package flickr_test

import (
	"os"
	"slices"
	"testing"

	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/config"
	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/flickr"
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

func TestFindUrls(t *testing.T) {
	body := `
	<a href="https://foo.com/bar"></a>
	<img src="http://foo.com/baz" />
	<script src="https://foo.com/qux"></script>
	`
	tests := map[string]struct {
		regex string
		want  []string
	}{
		"foo.com": {`https?://foo\.com/\w+`, []string{
			"https://foo.com/bar", "http://foo.com/baz", "https://foo.com/qux",
		}},
		"foo.net": {`https?://foo\.net/\w+`, []string{}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := config.Config{}
			c.Regex.Flickr.Url = tt.regex
			got := flickr.FindUrls(body, &c)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got: %#v, want: %#v\n", got, tt.want)
			}
		})
	}
}
func TestReplaceNeedlessTags(t *testing.T) {
	body := "<a href=foo>test</a><script src=foo></script>"
	tests := map[string]struct {
		reA, reScript, afterA, afterScript string
		want                               string
	}{
		"Normal case": {
			`<a href=foo>`,
			`<script src=foo></script>`,
			"<a href=bar>",
			"<script src=bar></script>",
			"<a href=bar>test</a><script src=bar></script>",
		},
		"Normal case with empty replace": {
			`<a href=foo>`,
			`<script src=foo></script>`,
			"<a tabindex=-1>",
			"",
			"<a tabindex=-1>test</a>",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := config.Config{}
			c.Regex.Flickr.Tag = map[string]string{}
			c.Regex.Flickr.Tag["a_start"] = tt.reA
			c.Regex.Flickr.Tag["script"] = tt.reScript
			c.Replace.Flickr.Tag = map[string]string{}
			c.Replace.Flickr.Tag["a_start"] = tt.afterA
			c.Replace.Flickr.Tag["script"] = tt.afterScript
			got := flickr.ReplaceNeedlessTags(body, &c)
			if got != tt.want {
				t.Errorf("got: %#v, want: %#v\n", got, tt.want)
			}
		})
	}
}
