package gormigrator

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"gorm.io/gorm"
)

const Version = "0.1.1-alpha"

var Testing bool = false

//Inspired by https://github.com/pressly/goose/blob/master/examples/go-migrations/00002_rename_root.go

type Store interface {
	GetCurrentLevel() (string, string, error)
	SaveState(code, level, user string) error
	FetchAll() ([]*Migration, error)
}

func InitMigration(db *gorm.DB) {
	//Get store
	store := NewMigrationStore(db)
	from, to, user, list, version := cmd()

	if *version {
		fmt.Println("Version: ", Version)
		return
	}

	if *list {
		err := showHistory(store)
		if err != nil {
			errorLog(err.Error())
		}
		return
	}

	if *from == "" {
		errorLog("no from-flag found")
	}
	if *to == "" {
		errorLog("no to-flag found")
	}

	err := StartMigration(from, to, user, store, db)
	if err != nil {
		errorLog(err.Error())
	}
}

func cmd() (*string, *string, *string, *bool, *bool) {
	cmd := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// migAction := cmd.String("mig", "status", "Migration mode up|down|status")
	from := cmd.String("from", "", "Current code")
	to := cmd.String("to", "", "Code to migrate to")
	user := cmd.String("user", "", "User who does migration")

	list := cmd.Bool("list", false, "List history of migration")
	version := cmd.Bool("version", false, "Show version")
	cmd.Parse(os.Args[1:])

	return from, to, user, list, version
}

func showHistory(store Store) error {

	list, _ := store.FetchAll()
	if len(list) == 0 {
		return errors.New("there is no migration done yet")
	}
	fmt.Printf("\n| %-40s | %-40s | %-20s |\n", "DATETIME HISTORY", "LEVEL (State)", "USER")
	fmt.Printf("| %-40s | %-40s | %-20s |\n", strings.Repeat("-", 40), strings.Repeat("-", 40), strings.Repeat("-", 20))
	for i, entry := range list {
		if i == 0 {
			fmt.Printf("| \033[;33m%-40s\033[0m | \033[;33m%-30s (current)\033[0m | \033[;33m%-20s\033[0m |\n", entry.CreatedAt, entry.Level, entry.User)
		} else {
			fmt.Printf("| %-40s | %-30s           | %-20s |\n", entry.CreatedAt, entry.Level, entry.User)
		}
	}
	return nil
}

func StartMigration(from, to, user *string, store Store, db *gorm.DB) error {

	consistencyFileCheck(&migrationFileList)

	// Convert data from files to an ordered list
	ordered := NewMigrationOrdered()
	ordered.importMigrationFileList(migrationFileList)

	if len(ordered) == 0 {
		return errors.New("no migration file found")
	}

	return doExecution(*from, *to, *user, store, &ordered, db)

}

func getIndices(from, to string, ordered *migrationOrderedType) (int, int, error) {
	startIndex, err := ordered.getIndex(from)
	if err != nil {
		return 0, 0, fmt.Errorf("from-code: %w", err)
	}
	endIndex, err := ordered.getIndex(to)
	if err != nil {
		return 0, 0, fmt.Errorf("to-code: %w", err)
	}
	return startIndex, endIndex, nil

}

func doExecution(from, to, user string, store Store, ordered *migrationOrderedType, db *gorm.DB) error {

	// Need to check if "from" is identical with current code from database, and it shouldn't be null
	_, code, err := store.GetCurrentLevel()
	if from == "null" && !errors.Is(err, gorm.ErrRecordNotFound) && code != "null" {
		return errors.New("there is a current state available")
	}

	if from != "null" {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("current code is not available")
		}
		if from != code {
			return errors.New("from-code not equal to state code")
		}
		if err != nil {
			return err
		}
	}

	start, end, err := getIndices(from, to, ordered)
	if err != nil {
		return err
	}

	err = ordered.execute(start, end, user, db, store)
	return err
}
