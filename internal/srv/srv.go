// Package srv is for service
package srv

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/ad9311/ninete/internal/errs"
)

type Store struct {
	conf conf
}

type conf struct {
	jwtSecret   string
	jwtIssuer   string
	jwtAudience []string
}

func New(db *sql.DB) (*Store, error) {
	var store *Store

	c, err := loadConf()
	if err != nil {
		return store, err
	}

	store = &Store{
		conf: c,
	}

	return store, nil
}

func loadConf() (conf, error) {
	var c conf

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return c, fmt.Errorf("%w: JWT_SECRET", errs.ErrEnvNoTSet)
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		return c, fmt.Errorf("%w: JWT_ISSUER", errs.ErrEnvNoTSet)
	}

	jwtAudience, err := loadList("JWT_AUDIENCE")
	if err != nil {
		return c, err
	}

	c = conf{
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtAudience: jwtAudience,
	}

	return c, nil
}

func loadList(envName string) ([]string, error) {
	var list []string

	str := os.Getenv(envName)
	if str == "" {
		return list, fmt.Errorf("%w: %s", errs.ErrEnvNoTSet, envName)
	}

	list = strings.Split(str, ",")

	return list, nil
}
