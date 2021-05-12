package gormigrator

import (
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

type updown struct {
	level string
	code  string
	up    func(*gorm.DB) error
	down  func(*gorm.DB) error
}

// The goal is to end up with an ordered list according to the filenames (level)
type migrationOrderedType []*updown

func NewMigrationOrdered() migrationOrderedType {
	return migrationOrderedType{}
}

// Convert file list to slice and order it by name
func (mig *migrationOrderedType) importMigrationFileList(m map[string]*updown) {
	slices := make([]*updown, 0, len(m))
	for _, updownEntry := range m {
		slices = append(slices, updownEntry)
	}
	sort.Slice(slices, func(i, j int) bool {
		return slices[i].level < slices[j].level
	})
	*mig = slices
}

func (mig *migrationOrderedType) getIndex(code string) (int, error) {
	if code == "null" {
		return -1, nil
	}
	for i, entry := range *mig {
		if entry.code == code {
			return i, nil
		}
	}
	return 0, errors.New("couldn't find code")
}

func (mig *migrationOrderedType) execute(startIndex int, endIndex int, db *gorm.DB, store Store) error {
	// Upgrade
	if startIndex < endIndex {
		for i := startIndex; i < endIndex; i++ {
			fmt.Println("\033[;33m⬆ UPGRADE ("+mig.getLevelOnIndex(i)+") ⟶ ("+mig.getLevelOnIndex(i+1)+") ", mig.getCodeOnIndex(i), "⟶", mig.getCodeOnIndex(i+1), "\033[0m")
			// execute upgrade
			err := (*mig)[i+1].up(db)
			if err == nil {
				if errStore := store.SaveState(mig.getCodeOnIndex(i+1), mig.getLevelOnIndex(i+1)); errStore != nil {
					return errStore
				}
			} else {
				fmt.Println("\033[;31mUPGRADE ERROR keep current state: ", mig.getCodeOnIndex(i), "\033[0m")
				fmt.Println("\033[;31mCheck your database and migration file\033[0m")
				return fmt.Errorf("%w at target state: %s (%s)", err, mig.getCodeOnIndex(i+1), mig.getLevelOnIndex(i+1))
			}
		}
	}

	// Downgrade
	if startIndex > endIndex {
		// Only one step down is allowed. This is to prevent accidentally downgrade to null state.
		if startIndex-endIndex > 1 {
			return errors.New("can't downgrade more than one step")
		}

		for i := startIndex; i > endIndex; i-- {
			fmt.Println("\033[;36m⬇ DOWNGRADE ("+mig.getLevelOnIndex(i)+") ⟶ ("+mig.getLevelOnIndex(i-1)+") ", mig.getCodeOnIndex(i), "⟶", mig.getCodeOnIndex(i-1), "\033[0m")
			err := (*mig)[i].down(db)
			if err == nil {
				if errStore := store.SaveState(mig.getCodeOnIndex(i-1), mig.getLevelOnIndex(i-1)); errStore != nil {
					return errStore
				}
			} else {
				fmt.Println("\033[;31mDOWNGRADE ERROR keep current state: ", mig.getCodeOnIndex(i), "\033[0m")
				fmt.Println("\033[;31mCheck your database and migration file\033[0m")
				return fmt.Errorf("%w at target state: %s (%s)", err, mig.getCodeOnIndex(i-1), mig.getLevelOnIndex(i-1))
			}
		}
	}

	return nil
}

func (mig *migrationOrderedType) getCodeOnIndex(index int) string {
	if index < 0 {
		return "null"
	} else {
		return (*mig)[index].code
	}
}

func (mig *migrationOrderedType) getLevelOnIndex(index int) string {
	if index < 0 {
		return "null"
	} else {
		return (*mig)[index].level
	}
}
