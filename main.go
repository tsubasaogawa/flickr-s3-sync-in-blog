package main

import (
	"bytes"
	"context"
	"flag"
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

var (
	bucket, entryPath, key, region string
)

func init() {
	flag.StringVar(&bucket, "bucket", "", "Upload S3 bucket name")
	flag.StringVar(&key, "key", "", "Upload S3 key")
	flag.StringVar(&region, "region", "ap-northeast-1", "Upload S3 region name")
}

func main() {
	flag.Parse()
	if bucket == "" || region == "" {
		log.Fatal("required args must be not empty")
	}

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
	imgName := filepath.Base(img_url[0])

	// up to s3
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Couldn't load default configuration. Have you set up your AWS account?")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	resp, err := http.Get(img_url[0])
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	// To calculate image size, should read data to memory
	// (Flickr cannot return Content-Length header).
	imgb, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if key != "" && !strings.HasSuffix(key, "/") {
		key += "/"
	}
	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key + imgName),
		Body:          bytes.NewBuffer(imgb),
		ContentLength: aws.Int64(int64(len(imgb))),
	}, s3.WithAPIOptions(v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware))
	if err != nil {
		log.Fatal(err)
	}
	// replace old flickr url to new s3 one
	// remove unused flickr attributes
}
