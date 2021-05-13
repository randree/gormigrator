package gormigrator

import (
	"gorm.io/gorm"
)

type MigrationStore struct {
	db *gorm.DB
}

func NewMigrationStore(db *gorm.DB) *MigrationStore {

	startMig := &Migration{}

	// Creating a migration table if not exists
	db.AutoMigrate(startMig)

	return &MigrationStore{
		db: db,
	}

}

// GetCurrentLevel returns the current level. In our case it is the filename, e.g. "mig0032.go"
func (m *MigrationStore) GetCurrentLevel() (string, string, error) {
	last := &Migration{}
	if err := m.db.Last(&last).Error; err != nil {
		return "", "", err
	}
	return last.Level, last.Code, nil
}

func (m *MigrationStore) FetchAll() ([]*Migration, error) {
	migrationList := make([]*Migration, 0)
	if err := m.db.Table("migrations").Order("id DESC").Find(&migrationList).Error; err != nil {
		return nil, err
	}
	return migrationList, nil
}

func (m *MigrationStore) SaveState(code, level, user string) error {
	currentState := &Migration{
		Code:  code,
		Level: level,
		User:  user,
	}
	if err := m.db.Create(&currentState).Error; err != nil {
		return err
	}
	return nil
}
