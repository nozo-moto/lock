package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestWorker(t *testing.T) {
	db, err := sql.Open("mysql", "root:pass@tcp(localhost:3306)/test")
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	defer db.Close()

	worker := NewWorker(
		db,
	)
	t.Run("main", func(t *testing.T) {
		err := worker.run(context.Background(), time.Now(), 1)
		if err != nil {
			t.Fatal(err)
		}
	})
}
