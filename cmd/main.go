package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/sync/errgroup"

	"github.com/tsubasaogawa/flickr-s3-sync/internal/entry"
	"github.com/tsubasaogawa/flickr-s3-sync/internal/flickr"
	"github.com/tsubasaogawa/flickr-s3-sync/internal/url"
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
	flag.IntVar(&threadLimit, "threadLimit", 2, "Limits for image download/upload threads")
}

func validation(bucket, region, path string) {
	if bucket == "" || region == "" {
		log.Fatal("required args must be not empty")
	}

	if entryPath == "" {
		log.Fatal("Arg is required")
	}
}

func main() {
	flag.Parse()

	entryPath = flag.Arg(0)
	validation(bucket, region, entryPath)

	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Couldn't load default configuration. Have you set up your AWS account?")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	// read entry text
	entry, err := entry.NewEntry(entryPath, backupDir, dryrun)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Load: " + entryPath)

	// pick up flickr image urls
	flickrImageUrls := entry.FindFlickrUrls()
	if flickrImageUrls == nil {
		log.Println("Entry has no Flickr url")
		return
	}

	var eg errgroup.Group
	eg.SetLimit(threadLimit)
	replaceUrlPairs := make(url.Urls, len(flickrImageUrls))

	// handle as flickr url
	for i, url := range flickrImageUrls {
		key := filepath.Join(dir, filepath.Base(url))

		// save old flickr url to new s3 one
		replaceUrlPairs[i].Old = url
		replaceUrlPairs[i].New = "https://" + bucket + "/" + key

		if dryrun || !uploadS3 {
			continue
		}

		// upload using goroutine
		u := url
		eg.Go(func() error {
			err = flickr.NewFlickr(u).CopyImageToS3(s3Client, bucket, key, overwrite)
			if err == nil {
				log.Println("Upload: " + key)
				return nil
			} else if errors.Is(err, os.ErrExist) {
				log.Println("Avoid overwriting: " + key)
				return nil
			}
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err.Error())
	}

	entry.Replace(replaceUrlPairs)

	if dryrun {
		fmt.Println(entry.NewBody)
		return
	}

	if backupDir != "" {
		backupFile, err := entry.Backup(entryPath, backupDir)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Backup: " + backupFile)
	}

	entry.Save()
	log.Println("Save: " + filepath.Clean(entryPath))
}
