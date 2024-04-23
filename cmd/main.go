package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	goConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"golang.org/x/sync/errgroup"

	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/config"
	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/entry"
	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/flickr"
	"github.com/tsubasaogawa/flickr-s3-sync-in-blog/internal/url"
)

var (
	confPath         string
	uploadS3, dryrun bool
)

func init() {
	flag.StringVar(&confPath, "config", "", "Configuration toml file path")
	flag.BoolVar(&uploadS3, "uploadS3", true, "Skip uploading to S3 when false")
	flag.BoolVar(&dryrun, "dryrun", false, "Dry run")
}

func setup() (string, *config.Config, *s3.Client, error) {
	flag.Parse()

	entryPath := flag.Arg(0)
	return setupUsingArg(entryPath)
}

func setupUsingArg(entryPath string) (string, *config.Config, *s3.Client, error) {
	if err := validation(entryPath); err != nil {
		return "", nil, nil, err
	}

	conf, err := config.NewConfig(confPath)
	if err != nil {
		return "", nil, nil, err
	}

	sdkConfig, err := goConf.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Couldn't load default configuration. Have you set up your AWS account?")
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	return entryPath, conf, s3Client, nil
}

func validation(entryPath string) error {
	if confPath == "" {
		return fmt.Errorf("confPath (fssync.toml) must be not empty")
	}

	if entryPath == "" {
		return fmt.Errorf("Argument <entryPath> is required. For example: /foo/bar/blog/posts.md")
	}

	return nil
}

func main() {
	entryPath, conf, s3c, err := setup()
	if err != nil {
		log.Fatal(err)
	}

	// read entry text
	entry, err := entry.NewEntry(entryPath, conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Load: " + entryPath)

	// pick up flickr image urls
	flickrImageUrls := flickr.FindUrls(entry.Body, conf)
	if flickrImageUrls == nil {
		log.Println("Entry has no Flickr url")
		return
	}

	var eg errgroup.Group
	eg.SetLimit(conf.General.ThreadLimit)
	replaceUrlPairs := make(url.Urls, len(flickrImageUrls))

	// handle as flickr url
	for i, url := range flickrImageUrls {
		key := filepath.Join(conf.S3.Directory, filepath.Base(url))

		// save old flickr url to new s3 one
		replaceUrlPairs[i].Old = url
		replaceUrlPairs[i].New = "https://" + conf.S3.Bucket + "/" + key

		if dryrun || !uploadS3 {
			continue
		}

		// upload using goroutine
		u := url
		eg.Go(func() error {
			err = flickr.NewFlickr(u, conf).CopyImageToS3(s3c, conf.S3.Bucket, key, conf.S3.Overwrite)
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

	if conf.General.BackupDir != "" {
		backupFile, err := entry.Backup(entryPath, conf.General.BackupDir)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Backup: " + backupFile)
	}

	entry.Save()
	log.Println("Save: " + filepath.Clean(entryPath))
}
