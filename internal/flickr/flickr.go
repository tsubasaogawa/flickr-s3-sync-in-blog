package flickr

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"regexp"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/tsubasaogawa/flickr-s3-sync-from-blog/internal/config"
)

type Flickr struct {
	Url    string
	config *config.Config
}

var (
	ReATag, ReScriptTag *regexp.Regexp
)

func NewFlickr(url string, conf *config.Config) *Flickr {
	ReATag = regexp.MustCompile(conf.Regex.Flickr.Tag["a"])
	ReScriptTag = regexp.MustCompile(conf.Regex.Flickr.Tag["script"])

	return &Flickr{
		Url:    url,
		config: conf,
	}
}

func (flickr *Flickr) CopyImageToS3(s3c *s3.Client, bucket, key string, overwrite bool) error {
	imgb, err := flickr.getImageByteData()
	if err != nil {
		return err
	}

	if !overwrite {
		list, err := s3c.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(key),
		})
		if err != nil {
			return err
		} else if *list.KeyCount > 0 {
			return os.ErrExist
		}
	}

	_, err = s3c.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          bytes.NewBuffer(*imgb),
		ContentLength: aws.Int64(int64(len(*imgb))),
	}, s3.WithAPIOptions(v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware))

	return err
}

func (flickr *Flickr) getImageByteData() (*[]byte, error) {
	// TODO: User-Agent
	resp, err := http.Get(flickr.Url)
	if err != nil {
		return &[]byte{}, err
	}
	defer resp.Body.Close()

	// To calculate image size, the program should read full data to memory
	// because Flickr cannot return Content-Length header.
	imgb, err := io.ReadAll(resp.Body)
	if err != nil {
		return &[]byte{}, err
	}

	return &imgb, nil
}

func FindUrls(body string, conf *config.Config) []string {
	return regexp.MustCompile(conf.Regex.Flickr.Url).FindAllString(body, -1)
}
