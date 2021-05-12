package gormigrator

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStartMigration(t *testing.T) {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open("host=localhost user=user password=passpass dbname=testdb port=5432 sslmode=disable"), &gorm.Config{Logger: newLogger})

	// In case of an error clean up test tables
	db.Migrator().DropTable("migrations")
	db.Migrator().DropTable("migtest")

	t.Run("DB ok", func(t *testing.T) {
		assert.NoError(t, err)
	})

	for _, test := range []struct {
		TestLabel    string
		Args         []string
		Output       string
		hasPanic     bool
		errorMessage string
	}{
		{
			TestLabel:    "Error no flag at all",
			Args:         []string{"cmd"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "no from-flag found",
		},
		{
			TestLabel:    "Error no to-flag set",
			Args:         []string{"cmd", "-from", "null"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "no to-flag found",
		},
		{
			TestLabel:    "Error no from-flag set",
			Args:         []string{"cmd", "-to", "user_table_start"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "no from-flag found",
		},
		{
			TestLabel: "Do first migration",
			Args:      []string{"cmd", "-from", "null", "-to", "user_table_start"},
			Output:    "21\n",
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat first migration",
			Args:         []string{"cmd", "-from", "null", "-to", "user_table_start"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "there is a current state available",
		},
		{
			TestLabel: "Do second migration",
			Args:      []string{"cmd", "-from", "user_table_start", "-to", "user_table_add_column"},
			Output:    "21\n",
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat second migration",
			Args:         []string{"cmd", "-from", "user_table_start", "-to", "user_table_add_column"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "from-code not equal to state code",
		},
		{
			TestLabel:    "Do third migration with fake database error",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_through_error"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "fake error at target state: user_table_through_error (mig0011.go)",
		},
		{
			TestLabel:    "Do first downgrade with error",
			Args:         []string{"cmd", "-from", "user_table_through_error", "-to", "user_table_add_column"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "from-code not equal to state code", // because it didn't upgrade before
		},
		{
			TestLabel:    "Do downgrade down to null",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "null"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "can't downgrade more than one step", // it's forbidden to go down more than one step
		},
		{
			TestLabel: "Do second downgrade",
			Args:      []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_start"},
			Output:    "21\n",
			hasPanic:  false,
		},
		{
			TestLabel:    "Repeat second downgrade",
			Args:         []string{"cmd", "-from", "user_table_add_column", "-to", "user_table_start"},
			Output:       "21\n",
			hasPanic:     true,
			errorMessage: "from-code not equal to state code", // because it didn't upgrade before
		},
		{
			TestLabel: "Final downgrade to null",
			Args:      []string{"cmd", "-from", "user_table_start", "-to", "null"},
			Output:    "21\n",
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

	// db.Migrator().DropTable("migrations")
	// Instead of dropping table you can investigate migrations
}
