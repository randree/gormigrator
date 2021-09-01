package gormigrator

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type State struct {
	Tag   string
	Level string
	Up    func(*gorm.DB) error
	Down  func(*gorm.DB) error
}

var migrations = []State{{Tag: "null", Level: "0"}}

func Mig(state State) {

	// 1. Check Tag
	filename := caller()
	if state.Tag == "" {
		log.Fatal("tag is missing (" + filename + ")")
	}
	if state.Tag == "null" {
		log.Fatal("tag can't have reserved name null (" + filename + ")")
	}
	if strings.Contains(state.Tag, " ") {
		log.Fatal("tag contains a whitespace (" + filename + ")")
	}
	if tagExists(state.Tag) {
		log.Fatal("tag already exists (" + filename + ")")
	}

	// 2. Check UP
	if state.Up == nil {
		log.Fatal("up function missing (" + filename + ")")
	}

	// 3. Check DOWN
	if state.Up == nil {
		log.Fatal("down function missing (" + filename + ")")
	}

	// Adding filename as level
	// in case Level is not set manually
	if state.Level == "" {
		state.Level = filename
	}

	migrations = append(migrations, state)

	// Sort after filename aka level
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Level < migrations[j].Level
	})
}

func tagExists(newTag string) bool {
	for _, migration := range migrations {
		if migration.Tag == newTag {
			return true
		}
	}
	return false
}

func caller() string {
	_, path, _, _ := runtime.Caller(2)
	return filepath.Base(path)
}

func performMigration(fromTag, toTag, user string, db *gorm.DB, migrationStore *MigrationStore) error {

	// No username
	if user == "" {
		return fmt.Errorf("no username set")
	}

	// Get current tag from db
	currentTag, err := migrationStore.GetCurrentTag()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// If there is no migration yet current tag will be empty thus "null"
	if currentTag == "" {
		currentTag = "null"
	}

	if currentTag == toTag && currentTag != "null" {
		fmt.Println("\033[;32mnothing to migrate - DB up-to-date\033[0m")
		return nil
	}

	if currentTag != fromTag {
		return fmt.Errorf("current tag differs from FROM tag")
	}

	if !tagExists(toTag) && toTag != "null" {
		return fmt.Errorf("TO tag does not exist")
	}
	// Find two points
	// The index for "null" will be 0 because it's not on the list
	var fromIndex, toIndex int
	for i, migration := range migrations {
		if migration.Tag == fromTag {
			fromIndex = i
		}
		if migration.Tag == toTag {
			toIndex = i
		}
	}

	for i := fromIndex + 1; i <= toIndex; i++ {
		fmt.Println("\033[;33m⬆ UPGRADE ("+migrations[i-1].Level+") ⟶ ("+migrations[i].Level+") tag:", migrations[i-1].Tag, "⟶", migrations[i].Tag, "\033[0m")
		err := migrations[i].Up(db)
		if err != nil {
			return err
		}
		migrationStore.SaveState(migrations[i].Tag, migrations[i].Level, user)
		if err != nil {
			return err
		}
	}

	// Only one step down by a time is allowed
	if toIndex-fromIndex < -1 {
		return fmt.Errorf("two steps DOWN is not allowed")
	}

	for i := fromIndex; i > toIndex; i-- {
		fmt.Println("\033[;36m⬇ DOWNGRADE ("+migrations[i].Level+") ⟶ ("+migrations[i-1].Level+") tag:", migrations[i].Tag, "⟶", migrations[i-1].Tag, "\033[0m")
		err := migrations[i].Down(db)
		if err != nil {
			return err
		}
		migrationStore.SaveState(migrations[i-1].Tag, migrations[i-1].Level, user)
		if err != nil {
			return err
		}
	}
	return nil
}
