package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

type Url struct {
	old string
	new string
}

var (
	bucket, entryPath, dir, region string
	overwrite, uploadS3            bool
)

func init() {
	flag.StringVar(&bucket, "s3Bucket", "", "Upload S3 bucket name")
	flag.StringVar(&dir, "s3Dir", "", "Upload S3 directory key")
	flag.StringVar(&region, "s3Region", "ap-northeast-1", "Upload S3 region name")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite when the photo has been already uploaded")
	flag.BoolVar(&uploadS3, "uploadS3", true, "Skip uploading to S3 when false")
}

func main() {
	flag.Parse()
	if bucket == "" || region == "" {
		log.Fatal("required args must be not empty")
	}

	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Couldn't load default configuration. Have you set up your AWS account?")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	rFlickrImageUrl := regexp.MustCompile(`https?://\w+\.staticflickr\.com/[0-9a-zA-Z_/]+\.(?:jpg|jpeg|png|gif)`)

	entryPath = flag.Arg(0)
	if entryPath == "" {
		log.Fatal("Arg is required")
	}

	entryTextb, err := os.ReadFile(entryPath)
	if err != nil {
		log.Fatal(err)
	}
	entryText := string(entryTextb)

	flickrImageUrls := rFlickrImageUrl.FindAllString(entryText, -1)
	if flickrImageUrls == nil {
		log.Println("Flickr url is not in an entry")
		os.Exit(0)
	}

	replaceUrlPairs := make([]Url, len(flickrImageUrls))
	for i, url := range flickrImageUrls {
		// up to s3
		imgb, err := getImageByteFromFlickr(url)
		if err != nil {
			log.Fatal(err)
		}

		if dir != "" && !strings.HasSuffix(dir, "/") {
			dir += "/"
		}
		key := dir + filepath.Base(url)
		if err = uploadToS3(s3Client, key, imgb); err != nil {
			log.Fatal(err)
		}

		// replace old flickr url to new s3 one
		replaceUrlPairs[i].old = url
		replaceUrlPairs[i].new = "https://" + bucket + "/" + key
	}
	fmt.Print(replaceFlickrHtml(parseUrl(replaceUrlPairs), entryText))
}

func getImageByteFromFlickr(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	// To calculate image size, the program should read full data to memory
	// because Flickr cannot return Content-Length header.
	imgb, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return imgb, nil
}

func uploadToS3(s3c *s3.Client, key string, imgb []byte) error {
	if !uploadS3 {
		return nil
	}
	list, err := s3c.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return err
	} else if !overwrite || *list.KeyCount > 0 {
		return nil
	}

	_, err = s3c.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewBuffer(imgb),
		ContentLength: aws.Int64(int64(len(imgb))),
	}, s3.WithAPIOptions(v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware))

	return err
}

func replaceFlickrHtml(replaceUrlPairs []string, text string) string {
	rFlickrATag := regexp.MustCompile(`<a.*href="https?://www\.flickr\.com/(?:photos/\w+/\d+/in/[^"]+|gp/\w+/\w+)"[^>]*>`)
	rFlickrScriptTag := regexp.MustCompile(`<script.*src="//embedr.flickr.com/assets/client-code.js"[^>]*></script>`)

	replacer := strings.NewReplacer(replaceUrlPairs...)
	return replacer.Replace(
		rFlickrScriptTag.ReplaceAllString(
			rFlickrATag.ReplaceAllString(text, `<a tabindex="-1">`),
			"",
		),
	)
}

func parseUrl(urls []Url) []string {
	parsed := make([]string, len(urls)*2)
	for _, url := range urls {
		parsed = append(parsed, url.old, url.new)
	}
	return parsed
}
