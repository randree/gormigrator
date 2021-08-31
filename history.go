package gormigrator

import (
	"errors"
	"fmt"
	"strings"
)

func showHistory(store *MigrationStore) error {

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
