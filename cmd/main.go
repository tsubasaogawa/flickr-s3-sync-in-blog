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

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"golang.org/x/sync/errgroup"

	"github.com/tsubasaogawa/hatenablog-flickr-to-s3-converter/internal/entry"
	"github.com/tsubasaogawa/hatenablog-flickr-to-s3-converter/internal/url"
)

var (
	// TODO: external file
	rFlickrImageUrl = regexp.MustCompile(`https?://\w+\.staticflickr\.com/[0-9a-zA-Z_/]+\.(?:jpg|jpeg|png|gif)`)
)

var (
	bucket, entryPath, dir, region, backupDir string
	overwrite, uploadS3, dryrun               bool
	threadLimit                               int
)

func init() {
	flag.StringVar(&bucket, "s3Bucket", "", "Upload S3 bucket name")
	flag.StringVar(&dir, "s3Dir", "", "Upload S3 directory key")
	flag.StringVar(&region, "s3Region", "ap-northeast-1", "Upload S3 region name")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite when the photo has been already uploaded")
	flag.BoolVar(&uploadS3, "uploadS3", true, "Skip uploading to S3 when false")
	flag.BoolVar(&dryrun, "dryrun", false, "Dry run")
	flag.StringVar(&backupDir, "backupDir", "", "Backup directory for an entry file")
	flag.IntVar(&threadLimit, "threadLimit", 4, "Limits for image download/upload threads")
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

	entryPath = flag.Arg(0)
	if entryPath == "" {
		log.Fatal("Arg is required")
	}

	// read entry text
	entry, err := entry.NewEntry(entryPath, backupDir, dryrun)
	if err != nil {
		log.Fatal(err)
	}

	// pick up flickr image urls
	flickrImageUrls := rFlickrImageUrl.FindAllString(entry.Body, -1)
	if flickrImageUrls == nil {
		log.Println("Flickr url is not in an entry")
		os.Exit(0)
	}

	replaceUrlPairs := make(url.Urls, len(flickrImageUrls))
	var eg errgroup.Group
	eg.SetLimit(threadLimit)

	for i, url := range flickrImageUrls {
		key := filepath.Join(dir, filepath.Base(url))
		if !dryrun {
			eg.Go(func() error {
				return copyFlickrImageToS3(s3Client, key, url)
			})
		}
		// replace old flickr url to new s3 one
		replaceUrlPairs[i].Old = url
		replaceUrlPairs[i].New = "https://" + bucket + "/" + key
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err.Error())
	}

	entry.Replace(replaceUrlPairs)
	if dryrun {
		fmt.Print(entry.Body)
		return
	}
	entry.Save()
}

func copyFlickrImageToS3(s3c *s3.Client, key, url string) error {
	if !uploadS3 {
		return nil
	}

	imgb, err := getImageByteData(url)
	if err != nil {
		return err
	}

	list, err := s3c.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return err
	} else if *list.KeyCount > 0 && !overwrite {
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

func getImageByteData(url string) ([]byte, error) {
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
