package gormigrator

import (
	"gorm.io/gorm"
)

// Migration model
type Migration struct {
	gorm.Model
	Tag   string `gorm:"size:455"`
	Level string `gorm:"size:255"`
	User  string `gorm:"size:255"`
}

type MigrationStore struct {
	tableName string
	db        *gorm.DB
}

func NewMigrationStore(db *gorm.DB, tableName string) *MigrationStore {

	if tableName == "" {
		tableName = "migrations"
	}

	// Creating a migration table if not exists
	db.Table(tableName).AutoMigrate(&Migration{})

	return &MigrationStore{
		tableName: tableName,
		db:        db,
	}

}

// GetCurrentLevel returns the current level. In our case it is the filename, e.g. "mig0032.go"
func (m *MigrationStore) GetCurrentTag() (string, error) {
	last := &Migration{}
	if err := m.db.Table(m.tableName).Last(&last).Error; err != nil {
		return "", err
	}
	return last.Tag, nil
}

func (m *MigrationStore) FetchAll() ([]*Migration, error) {
	var migrationList []*Migration
	if err := m.db.Table(m.tableName).Order("id DESC").Find(&migrationList).Error; err != nil {
		return nil, err
	}
	return migrationList, nil
}

func (m *MigrationStore) SaveState(tag, level, user string) error {
	currentState := &Migration{
		Tag:   tag,
		Level: level,
		User:  user,
	}
	if err := m.db.Table(m.tableName).Create(&currentState).Error; err != nil {
		return err
	}
	return nil
}
