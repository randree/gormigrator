package gormigrator

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gorm.io/gorm"
)

//Inspired by https://github.com/pressly/goose/blob/master/examples/go-migrations/00002_rename_root.go

type Store interface {
	GetCurrentLevel() (string, string, error)
	SaveState(code, level string) error
}

func InitMigration(db *gorm.DB) {
	//Get store
	store := NewMigrationStore(db)
	from, to := cmd()
	StartMigration(from, to, store, db)
}

func cmd() (*string, *string) {
	cmd := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// migAction := cmd.String("mig", "status", "Migration mode up|down|status")
	migFrom := cmd.String("from", "", "Current code")
	migTo := cmd.String("to", "", "Code to migrate to")
	cmd.Parse(os.Args[1:])

	// if *migAction == "" {
	// 	panic("no migration flag found")
	// }
	if *migFrom == "" {
		panic("no from-flag found")
	}
	if *migTo == "" {
		panic("no to-flag found")
	}
	return migFrom, migTo
}

func StartMigration(from, to *string, store Store, db *gorm.DB) {

	consistencyFileCheck(&migrationFileList)

	// Convert data from files to an ordered list
	ordered := NewMigrationOrdered()
	ordered.importMigrationFileList(migrationFileList)

	if len(ordered) == 0 {
		panic("no migration file found")
	}

	err := doExecute(*from, *to, store, &ordered, db)

	if err != nil {
		panic(err.Error())

	}

}

func getIndices(from, to string, ordered *migrationOrderedType) (int, int, error) {
	startIndex, err := ordered.getIndex(from)
	if err != nil {
		return 0, 0, fmt.Errorf("StartIndex: %w", err)
	}
	endIndex, _ := ordered.getIndex(to)
	if err != nil {
		return 0, 0, fmt.Errorf("EndIndex: %w", err)
	}
	return startIndex, endIndex, nil

}

func doExecute(from, to string, store Store, ordered *migrationOrderedType, db *gorm.DB) error {

	// Need to check if "from" is identical with current code from database
	_, code, err := store.GetCurrentLevel()
	if from == "null" && !errors.Is(err, gorm.ErrRecordNotFound) {
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

	err = ordered.execute(start, end, db, store)
	return err
}
