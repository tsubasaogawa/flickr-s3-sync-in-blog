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
	bucket, domain, entryPath, dir, region string
	overwrite                              bool
)

func init() {
	flag.StringVar(&bucket, "s3Bucket", "", "Upload S3 bucket name")
	flag.StringVar(&dir, "s3Dir", "", "Upload S3 directory key")
	flag.StringVar(&region, "s3Region", "ap-northeast-1", "Upload S3 region name")
	flag.StringVar(&domain, "s3Domain", "", "S3 domain")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite when the photo has been already uploaded")
}

func main() {
	flag.Parse()
	if bucket == "" || region == "" {
		log.Fatal("required args must be not empty")
	}
	if domain == "" {
		domain = bucket
	}

	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Couldn't load default configuration. Have you set up your AWS account?")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	// r_a_url := regexp.MustCompile(`https?://www\.flickr\.com/gp/\w+/\w+`)
	// TODO: more extension
	rFlickrUrl := regexp.MustCompile(`https?://\w+\.staticflickr\.com/\w+/\w+\.jpg`)

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

	// a_url := r_a_url.FindStringSubmatch(article)
	flickrUrls := rFlickrUrl.FindAllString(entry, -1)
	if flickrUrls == nil {
		log.Fatal("Url is not found")
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
		} else if overwrite || *result.KeyCount < 1 {
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
		newUrl := "https://" + domain + "/" + key
		replaceUrlPairs[i*2] = url
		replaceUrlPairs[i*2+1] = newUrl

		// remove unused flickr attributes
	}
	replacer := strings.NewReplacer(replaceUrlPairs...)
	fmt.Print(replacer.Replace(entry))
}
