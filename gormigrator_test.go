package gormigrator

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// These file mocks can't be tested in real files because all calls are nested in the init part
// there are scenarios where they interfere with actual migration files

func MigrationFile0001Mock() {

	migrationFileList["mig0001.go"] = &updown{
		level: "mig0001.go",
		code:  "user_table_start",
		up: func(db *gorm.DB) error {
			type migtest struct {
				gorm.Model
				Name  string `gorm:"size:255"`
				Email string `gorm:"size:300"`
			}
			err := db.AutoMigrate(&migtest{})
			return err
		},
		down: func(db *gorm.DB) error {
			err := db.Migrator().DropTable("migtests")
			return err
		},
	}
}

func MigrationFile0002Mock() {

	migrationFileList["mig0002.go"] = &updown{
		level: "mig0002.go",
		code:  "user_table_add_column",
		up: func(db *gorm.DB) error {
			type migtest struct {
				// only new column in struct
				Testcol string
			}
			err := db.Migrator().AddColumn(&migtest{}, "testcol")
			return err
		},
		down: func(db *gorm.DB) error {
			type migtest struct {
				Testcol string
			}
			err := db.Migrator().DropColumn(&migtest{}, "testcol")
			return err
		},
	}
}

func MigrationFile0003Mock() {

	migrationFileList["mig00011.go"] = &updown{
		level: "mig0011.go",
		code:  "user_table_through_error",
		up: func(db *gorm.DB) error {
			err := errors.New("fake error")
			return err
		},
		down: func(db *gorm.DB) error {
			return nil
		},
	}
}

func TestStartMigration(t *testing.T) {

	MigrationFile0001Mock()
	MigrationFile0002Mock()
	MigrationFile0003Mock()

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: false,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open("host=localhost user=user password=passpass dbname=testdb port=5432 sslmode=disable"), &gorm.Config{Logger: newLogger})

	// In case of an error clean up test tables
	db.Migrator().DropTable("migrations")
	db.Migrator().DropTable("migtest")

	//activate testing
	Testing = true

	t.Run("DB ok", func(t *testing.T) {
		assert.NoError(t, err)
	})

	for _, test := range []struct {
		TestLabel    string
		Args         []string
		hasPanic     bool
		errorMessage string
	}{
		{
			TestLabel:    "Empty migration list error",
			Args:         []string{"cmd", "-list"},
			hasPanic:     true,
			errorMessage: "there is no migration done yet",
		},
		{
			TestLabel:    "Error no flag at all",
			Args:         []string{"cmd"},
			hasPanic:     true,
			errorMessage: "no from-flag found",
		},
		{
			TestLabel:    "Error no to-flag set",
			Args:         []string{"cmd", "-from", "null"},
			hasPanic:     true,
			errorMessage: "no to-flag found",
		},
		{
			TestLabel:    "Error no from-flag set",
			Args:         []string{"cmd", "-to", "user_table_start"},
			hasPanic:     true,
			errorMessage: "no from-flag found",
		},
		{
			TestLabel:    "Do first wrong migration",
			Args:         []string{"cmd", "-from", "null", "-to", "doNotExist", "-user", "John"},
			hasPanic:     true,
			errorMessage: "to-code: couldn't find code",
		},
		{
			TestLabel: "Do first migration",
			Args:      []string{"cmd", "-from", "null", "-to", "user_table_start", "-user", "Dan"},
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat first migration",
			Args:         []string{"cmd", "-from", "null", "-to", "user_table_start", "-user", "Dan"},
			hasPanic:     true,
			errorMessage: "there is a current state available",
		},
		{
			TestLabel: "Do second migration",
			Args:      []string{"cmd", "-from", "user_table_start", "-to", "user_table_add_column", "-user", "Maddy"},
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat second migration",
			Args:         []string{"cmd", "-from", "user_table_start", "-to", "user_table_add_column"},
			hasPanic:     true,
			errorMessage: "from-code not equal to state code",
		},
		{
			TestLabel:    "Do third migration with fake database error",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_through_error"},
			hasPanic:     true,
			errorMessage: "fake error at target state: user_table_through_error (mig0011.go)",
		},
		{
			TestLabel:    "Do first downgrade with error",
			Args:         []string{"cmd", "-from", "user_table_through_error", "-to", "user_table_add_column"},
			hasPanic:     true,
			errorMessage: "from-code not equal to state code", // because it didn't upgrade before
		},
		{
			TestLabel:    "Do downgrade down to null",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "null"},
			hasPanic:     true,
			errorMessage: "can't downgrade more than one step", // it's forbidden to go down more than one step
		},
		{
			TestLabel: "Do second downgrade",
			Args:      []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_start", "-user", "Team Migration"},
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat second downgrade",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_start", "-user", "Team Migration"},
			hasPanic:     true,
			errorMessage: "from-code not equal to state code", // because it didn't upgrade before
		},
		{
			TestLabel: "Final downgrade to null",
			Args:      []string{"cmd", "-from", "user_table_start", "-to", "null", "-user", "The Downgrader"},
			hasPanic:  false,
		},
		{
			TestLabel: "Do first migration again",
			Args:      []string{"cmd", "-from", "null", "-to", "user_table_start", "-user", "Dan2"},
			hasPanic:  false,
		},
		{
			TestLabel: "Final downgrade to null again",
			Args:      []string{"cmd", "-from", "user_table_start", "-to", "null", "-user", "The Downgrader2"},
			hasPanic:  false,
		},
		{
			TestLabel: "Show list of all migrations",
			Args:      []string{"cmd", "-list"},
			hasPanic:  false,
		},
		{
			TestLabel: "Show version",
			Args:      []string{"cmd", "-version"},
			hasPanic:  false,
		},
	} {
		t.Run(test.TestLabel, func(t *testing.T) {
			os.Args = test.Args
			if test.hasPanic {
				assert.PanicsWithValue(t, test.errorMessage, func() { InitMigration(db) })
			} else {
				InitMigration(db)
			}
		})
	}

	db.Migrator().DropTable("migrations")
	// Instead of dropping table you can investigate migrations
}
