package config

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
)

type (
	Config struct {
		General general
		S3      s3
		Regex   regex
		Replace replace
	}

	general struct {
		BackupDir         string        `toml:"backup_dir"`
		ThreadLimit       int           `toml:"thread_limit"`
		SleepSecForFlickr time.Duration `toml:"sleep_sec_for_flickr"`
	}

	s3 struct {
		Bucket, Directory, Region string
		Overwrite                 bool
	}

	regex struct {
		EntryPath map[string]string `toml:"entry_path"`
		Flickr    struct {
			Url string
			Tag map[string]string
		}
	}

	replace struct {
		Flickr struct {
			Tag map[string]string
		}
	}
)

func NewConfig(f string) (*Config, error) {
	c := &Config{}
	if _, err := toml.DecodeFile(f, c); err != nil {
		return nil, err
	}
	if c.hasInvalidValue() {
		return nil, fmt.Errorf("Invalid configuration: %#v", *c)
	} else if c.General.ThreadLimit > runtime.NumCPU() {
		c.General.ThreadLimit = runtime.NumCPU()
		log.Printf("thread_limit is changed to the #cpu (%d)\n", runtime.NumCPU())
	}

	return c, nil
}

func (c *Config) hasInvalidValue() bool {
	if c.S3.Bucket == "" || c.S3.Region == "" || c.Regex.Flickr.Url == "" {
		return true
	}

	hasAllKey := func(m map[string]string, keys []string) bool {
		for _, k := range keys {
			if _, ok := m[k]; !ok {
				return false
			}
		}
		return true
	}
	if !hasAllKey(c.Regex.Flickr.Tag, []string{"a_start", "script"}) {
		return true
	}

	return false
}
