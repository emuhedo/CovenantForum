// Package config provides the forum configuration file structure,
// initialization and reading.
package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/kelseyhightower/envconfig"
)

// Config is a forum configuration struct.
type Config struct {
	Address string `hcl:"address" envconfig:"BEBOP_ADDRESS"`
	BaseURL string `hcl:"base_url" envconfig:"BEBOP_BASE_URL"`
	Title   string `hcl:"title" envconfig:"BEBOP_TITLE"`

	JWT struct {
		Secret string `hcl:"secret" envconfig:"BEBOP_JWT_SECRET"`
	} `hcl:"jwt"`

	FileStorage struct {
		Type string `hcl:"type" envconfig:"BEBOP_FILE_STORAGE_TYPE"`

		Local struct {
			Dir string `hcl:"dir" envconfig:"BEBOP_FILE_STORAGE_LOCAL_DIR"`
		} `hcl:"local"`

		GoogleCloudStorage struct {
			ServiceAccountFile string `hcl:"service_account_file" envconfig:"BEBOP_FILE_STORAGE_GCS_SERVICE_ACCOUNT_FILE"`
			Bucket             string `hcl:"bucket" envconfig:"BEBOP_FILE_STORAGE_GCS_BUCKET"`
		} `hcl:"google_cloud_storage"`

		AmazonS3 struct {
			AccessKey string `hcl:"access_key" envconfig:"BEBOP_FILE_STORAGE_S3_ACCESS_KEY"`
			SecretKey string `hcl:"secret_key" envconfig:"BEBOP_FILE_STORAGE_S3_SECRET_KEY"`
			Region    string `hcl:"region" envconfig:"BEBOP_FILE_STORAGE_S3_REGION"`
			Bucket    string `hcl:"bucket" envconfig:"BEBOP_FILE_STORAGE_S3_BUCKET"`
		} `hcl:"amazon_s3"`

		CovenantS3 struct {
			Database  string `hcl:"database" envconfig:"BEBOP_STORE_COVENANTSQL_DATABASE"`
			Config    string `hcl:"config" envconfig:"BEBOP_STORE_COVENANTSQL_CONFIG"`
			MasterKey string `hcl:"masterkey" envconfig:"BEBOP_STORE_COVENANTSQL_MASTERKEY"`
		} `hcl:"covenant_s3"`

	} `hcl:"file_storage"`

	Store struct {
		Type string `hcl:"type" envconfig:"BEBOP_STORE_TYPE"`

		CovenantSQL struct {
			Database  string `hcl:"database" envconfig:"BEBOP_STORE_COVENANTSQL_DATABASE"`
			Config    string `hcl:"config" envconfig:"BEBOP_STORE_COVENANTSQL_CONFIG"`
			MasterKey string `hcl:"masterkey" envconfig:"BEBOP_STORE_COVENANTSQL_MASTERKEY"`
		} `hcl:"covenantsql"`
	} `hcl:"store"`

	OAuth struct {
		Google struct {
			ClientID string `hcl:"client_id" envconfig:"BEBOP_OAUTH_GOOGLE_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"BEBOP_OAUTH_GOOGLE_SECRET"`
		} `hcl:"google"`

		Facebook struct {
			ClientID string `hcl:"client_id" envconfig:"BEBOP_OAUTH_FACEBOOK_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"BEBOP_OAUTH_FACEBOOK_SECRET"`
		} `hcl:"facebook"`

		Github struct {
			ClientID string `hcl:"client_id" envconfig:"BEBOP_OAUTH_GITHUB_CLIENT_ID"`
			Secret   string `hcl:"secret" envconfig:"BEBOP_OAUTH_GITHUB_SECRET"`
		} `hcl:"github"`
	} `hcl:"oauth"`
}

// ReadFile reads a forum config from file.
func ReadFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	cfg := &Config{}
	err = hcl.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshal hcl: %v", err)
	}

	prepare(cfg)
	return cfg, nil
}

// ReadEnv reads a forum config from environment variables.
func ReadEnv() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %v", err)
	}
	prepare(cfg)
	return cfg, nil
}

func prepare(cfg *Config) {
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
}

// Init generates an initial config string.
func Init() (string, error) {
	buf := new(bytes.Buffer)
	err := tpl.Execute(buf, map[string]interface{}{
		"jwt_secret": GenKeyHex(32),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenKeyHex generates a crypto-random key with byte length byteLen
// and hex-encodes it to a string.
func GenKeyHex(byteLen int) string {
	bytes := make([]byte, byteLen)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

var tpl = template.Must(template.New("initial-config").Parse(strings.TrimSpace(`
address  = "127.0.0.1:8080"
base_url = "https://example.com/forum"
title    = "CovenantForum"

jwt {
  secret = "{{.jwt_secret}}"
}

file_storage {
  type = "covenant_s3"

  local {
    dir = "./"
  }

  covenant_s3 {
    database  = ""
    config    = ""
    masterkey = ""
  }
}

store {
  type = "covenantsql"

  covenantsql {
    database  = ""
    config    = ""
    masterkey = ""
  }
}

oauth {
  google {
    client_id = ""
    secret    = ""
  }

  facebook {
    client_id = ""
    secret    = ""
  }

  github {
    client_id = ""
    secret    = ""
  }
}
`)))
