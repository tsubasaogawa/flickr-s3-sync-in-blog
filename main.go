package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var (
	entryPath string
)

func init() {
}

func main() {
	flag.Parse()

	// r_a_url := regexp.MustCompile(`https?://www\.flickr\.com/gp/\w+/\w+`)
	r_img_url := regexp.MustCompile(`https?://\w+\.staticflickr\.com/\w+/\w+\.jpg`)

	entryPath = flag.Arg(0)
	if entryPath == "" {
		log.Fatal("Arg is required")
	}

	file, err := os.Open(entryPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	entryb, err := io.ReadAll(file)
	entry := string(entryb)

	// a_url := r_a_url.FindStringSubmatch(article)
	img_url := r_img_url.FindStringSubmatch(entry)
	// download
	download(img_url[0], "/tmp/"+filepath.Base(img_url[0]))
	// up to s3
	// replace old flickr url to new s3 one
	// remove unused flickr attributes
}

func download(remote, local string) {
	out, err := os.Create(local)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	resp, err := http.Get(remote)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
