package config

import (
	"fmt"
	"os"
)

type EnvKeys struct {
	Keys       []string
	DeployKeys []string
}

func NewEnvKeys() *EnvKeys {
	envs := []string{
		"SERVER_PORT", "COOKIE_NAME", "AWS_REGION",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"PROFILE_IMAGE_S3_BUCKET", "SES_EMAIL", "SENTRY_DSN",
	}
	deployKeys := []string{}
	deployKeys = append(deployKeys, "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_HOST", "POSTGRES_PORT")
	envKeys := EnvKeys{
		Keys:       envs,
		DeployKeys: deployKeys,
	}
	return &envKeys
}

type Env struct {
	//env keys list
	Keys EnvKeys

	//server envs
	GOENV       string
	SERVER_PORT string
	COOKIE_NAME string
	AWS_REGION  string
	SENTRY_DSN  string

	//aws access keys
	AWS_ACCESS_KEY_ID     string
	AWS_SECRET_ACCESS_KEY string

	//gcp oauth2 client secres
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string

	//s3 envs
	PROFILE_IMAGE_S3_BUCKET string

	//ses email
	SES_EMAIL string

	//dev db envs
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_DB       string
	POSTGRES_HOST     string
	POSTGRES_PORT     string
	DB_URI            string

	//prod RDS envs
	RDS_ENDPOINT       string
	RDS_USERNAME       string
	RDS_PASSWORD       string
	RDS_ENGINE         string
	RDS_ENGINE_VERSION string
}

func NewEnv() *Env {
	envKeys := NewEnvKeys()
	return &Env{
		Keys: *envKeys,
	}
}

func (e *Env) LookUpOSEnvs() error {
	keys := []string{}
	keys = append(keys, append(e.Keys.Keys, e.Keys.DeployKeys...)...)
	for _, v := range keys {
		if _, ok := os.LookupEnv(v); !ok {
			return fmt.Errorf("environment var %s doesn't exists", v)
		}
	}
	return nil
}

func (e *Env) InitOSEnv() *Env {
	err := e.LookUpOSEnvs()
	if err != nil {
		panic(err.Error())
	}
	e.SERVER_PORT = fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))
	e.COOKIE_NAME = os.Getenv("COOKIE_NAME")
	e.SENTRY_DSN = os.Getenv("SENTRY_DSN")

	e.AWS_ACCESS_KEY_ID = os.Getenv("AWS_ACCESS_KEY_ID")
	e.AWS_SECRET_ACCESS_KEY = os.Getenv("AWS_SECRET_ACCESS_KEY")

	e.GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
	e.GOOGLE_CLIENT_SECRET = os.Getenv("GOOGLE_CLIENT_SECRET")

	e.PROFILE_IMAGE_S3_BUCKET = os.Getenv("PROFILE_IMAGE_S3_BUCKET")
	e.SES_EMAIL = os.Getenv("SES_EMAIL")

	e.POSTGRES_USER = os.Getenv("POSTGRES_USER")
	e.POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	e.POSTGRES_DB = os.Getenv("POSTGRES_DB")
	e.POSTGRES_HOST = os.Getenv("POSTGRES_HOST")
	e.POSTGRES_PORT = os.Getenv("POSTGRES_PORT")

	e.RDS_ENDPOINT = os.Getenv("RDS_ENDPOINT")
	e.RDS_USERNAME = os.Getenv("RDS_USERNAME")
	e.RDS_PASSWORD = os.Getenv("RDS_PASSWORD")
	e.RDS_ENGINE = os.Getenv("RDS_ENGINE")
	e.RDS_ENGINE_VERSION = os.Getenv("RDS_ENGINE_VERSION")

	dbURI := os.Getenv("DATABASE_URL")
	e.DB_URI = dbURI
	return e
}
