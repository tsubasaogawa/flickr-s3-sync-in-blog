package main

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestSetupUsingArg(t *testing.T) {
	tests := map[string]struct {
		entryPath, confPath, uploadS3, dryrun string
		hasErr                                bool
	}{
		"confPath is empty":     {"foo", "", "true", "false", true},
		"confPath is not empty": {"foo", "foo", "true", "false", false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			flag.CommandLine.Set("config", tt.confPath)
			flag.CommandLine.Set("uploadS3", tt.uploadS3)
			flag.CommandLine.Set("dryrun", tt.dryrun)

			if _, _, _, err := setupUsingArg(tt.entryPath); tt.hasErr == (err == nil) {
				t.Errorf("got: %#v, want: %#v\n", err != nil, tt.hasErr)
			}
		})
	}
}
