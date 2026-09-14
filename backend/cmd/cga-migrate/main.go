package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kxs888/kxs/backend/internal/repo"
)

func main() {
	dir := flag.String("dir", "up", "up | down")
	flag.Parse()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required for migrate")
		os.Exit(1)
	}
	var err error
	switch *dir {
	case "down":
		err = repo.MigrateDown(url)
	default:
		err = repo.Migrate(url)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
