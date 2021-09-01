package gormigrator

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

const Version = "1.1.0"

var Testing bool = false

//Inspired by https://github.com/pressly/goose/blob/master/examples/go-migrations/00002_rename_root.go

func InitMigration(db *gorm.DB) {

	migrationStore := NewMigrationStore(db, "migrations")

	// If version show
	if getEnv("VERSION", "") == "1" {
		fmt.Println("Gormigrator version: ", Version)
	}

	// If history show
	if getEnv("HISTORY", "") == "1" {
		showHistory(migrationStore)
	}

	fromTag := getEnv("FROM", "")
	toTag := getEnv("TO", "")
	user := getEnv("USER", "")

	if fromTag == "" {
		fmt.Println("FROM tag is missing")
		showExample()
		return
	}
	if toTag == "" {
		fmt.Println("TO tag is missing")
		showExample()
		return
	}
	if user == "" {
		fmt.Println("no USER set")
		showExample()
		return
	}

	err := performMigration(fromTag, toTag, user, db, migrationStore)
	if err != nil {
		fmt.Println("\033[;31mMIGRATION ERROR! Keep last migration step\033[0m")
		fmt.Println(err)
		return
	}
}

func showExample() {
	fmt.Println("Examples:")
	fmt.Println("HISTORY=1 VERSION=1 FROM=foo TO=bar USER=tester go run ./...")
	fmt.Println("HISTORY=1 VERSION=1 FROM=foo TO=bar USER=tester  main_build")
}

// Get env variable as string
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
