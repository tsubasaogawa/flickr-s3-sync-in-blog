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

	rFlickrATag := regexp.MustCompile(`<a.*href="https?://www\.flickr\.com/(?:photos/\w+/\d+/in/[^"]+|gp/\w+/\w+)"[^>]*>`)
	rFlickrScriptTag := regexp.MustCompile(`<script.*src="//embedr.flickr.com/assets/client-code.js"[^>]*></script>`)
	rFlickrUrl := regexp.MustCompile(`https?://\w+\.staticflickr\.com/[0-9a-zA-Z_/]+\.(?:jpg|jpeg|png|gif)`)

	entryPath = flag.Arg(0)
	if entryPath == "" {
		log.Fatal("Arg is required")
	}

	f, err := os.Open(entryPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	entryb, err := io.ReadAll(f)
	entry := string(entryb)

	flickrUrls := rFlickrUrl.FindAllString(entry, -1)
	if flickrUrls == nil {
		log.Println("Flickr url is not found")
		os.Exit(0)
	}
	replaceUrlPairs := make([]string, len(flickrUrls)*2)
	for i, url := range flickrUrls {
		// up to s3
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		// To calculate image size, the program should read full data to memory
		// because Flickr cannot return Content-Length header.
		imgb, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if dir != "" && !strings.HasSuffix(dir, "/") {
			dir += "/"
		}
		key := dir + filepath.Base(url)
		result, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(key),
		})
		if err != nil {
			log.Fatal(err)
		} else if uploadS3 && (overwrite || *result.KeyCount < 1) {
			_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				Body:          bytes.NewBuffer(imgb),
				ContentLength: aws.Int64(int64(len(imgb))),
			}, s3.WithAPIOptions(v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware))
			if err != nil {
				log.Fatal(err)
			}
		}

		// replace old flickr url to new s3 one
		newUrl := "https://" + bucket + "/" + key
		replaceUrlPairs[i*2] = url
		replaceUrlPairs[i*2+1] = newUrl
	}
	replacer := strings.NewReplacer(replaceUrlPairs...)
	fmt.Print(replacer.Replace(
		rFlickrScriptTag.ReplaceAllString(
			rFlickrATag.ReplaceAllString(entry, `<a tabindex="-1">`),
			"",
		),
	))
}
