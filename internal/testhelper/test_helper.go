package testhelper

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ad9311/ninete/internal/db"
	"github.com/ad9311/ninete/internal/prog"
)

func OpenTestDB(url string) (*sql.DB, error) {
	var sqlDB *sql.DB

	sqlDB, err := sql.Open("sqlite3", "file:"+url+"?_loc=UTC")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping test database, %w", err)
	}

	return sqlDB, nil
}

func SetUpTestDB(dbName string) error {
	testDBDir := os.Getenv("TEST_DATABASE_DIR")

	if testDBDir == "" {
		return fmt.Errorf("TEST_DATABASE_DIR not set, %w", prog.ErrEnvNoTSet)
	}

	testDBURL := testDBDir + "/" + dbName

	if err := os.Setenv("DATABASE_URL", testDBURL); err != nil {
		return fmt.Errorf("failed to set DATABASE_URL, %w", err)
	}

	sqlDB, err := OpenTestDB(testDBURL)
	if err != nil {
		return fmt.Errorf("failed to open test database, %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("failed to close test database, %v", err)
		}
	}()

	return nil
}

func SetUpPackageTest(dbName string) int {
	if err := SetUpTestDB(dbName); err != nil {
		log.Printf("failed to set up test database, %v", err)

		return 1
	}

	if err := db.RunMigrationsUp(); err != nil {
		log.Printf("failed to run test migrations, %v", err)

		return 1
	}

	return 0
}

func SetJSONHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func SetAuthHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

func MarshalPayload(t *testing.T, params any) *bytes.Buffer {
	t.Helper()

	body, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal params, %v", err)
	}

	return bytes.NewBuffer(body)
}

func UnmarshalPayload(t *testing.T, res *httptest.ResponseRecorder, payload any) {
	t.Helper()

	if err := json.Unmarshal(res.Body.Bytes(), payload); err != nil {
		t.Fatalf("failed to unmarshal payload, %v", err)
	}
}
