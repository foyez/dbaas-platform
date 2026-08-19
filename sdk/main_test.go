package dbaas_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// Loads sdk/.env into the process's environment. Ignored if missing -
	// so CI can supply real env vars instead without needing a .env file.
	_ = godotenv.Load()

	os.Exit(m.Run())
}
