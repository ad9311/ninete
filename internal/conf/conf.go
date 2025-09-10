// Package conf
package conf

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// ENVs
const (
	ENVProduction  = "production"
	ENVDevelopment = "development"
	ENVTest        = "test"
)

// AppConf
type AppConf struct {
	ENV        string
	DBConf     DBConf
	ServerConf ServerConf
	Secrets    Secrets
}

// Load
func Load() (AppConf, error) {
	var ac AppConf

	env, err := loadENV()
	if err != nil {
		return ac, err
	}

	dbc, err := LoadDBConf()
	if err != nil {
		return ac, err
	}

	sc, err := LoadServerConf()
	if err != nil {
		return ac, err
	}

	scrt, err := LoadSecrets()
	if err != nil {
		return ac, err
	}

	ac = AppConf{
		ENV:        env,
		DBConf:     dbc,
		ServerConf: sc,
		Secrets:    scrt,
	}

	return ac, nil
}

func loadENV() (string, error) {
	env, ok := os.LookupEnv("NINETE_ENV")
	if !ok {
		return env, nil // ERROR
	}

	if err := isValidENV(env); err != nil {
		return "", err
	}

	if env != ENVProduction {
		path, ok := findRelativeENVFile()
		if err := godotenv.Load(path); !ok || err != nil {
			return "", err
		}
	}

	return env, nil
}

func isValidENV(env string) error {
	ok := validENVs()[env]
	if !ok {
		return nil // Error
	}

	return nil
}

func validENVs() map[string]bool {
	return map[string]bool{
		ENVProduction:  true,
		ENVDevelopment: true,
		ENVTest:        true,
	}
}

func findRelativeENVFile() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		p := filepath.Join(dir, ".env")
		if fileExists(p) {
			return p, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func fileExists(p string) bool {
	_, err := os.Stat(p)

	return err == nil
}
