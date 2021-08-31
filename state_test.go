package gormigrator

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createMockMigrationList() {

	type migtest struct {
		ID    string `gorm:"primarykey"`
		Name  string `gorm:"size:255"`
		Email string `gorm:"size:300"`
	}

	Mig(State{
		Level: "mig001",
		Tag:   "create_migtest",
		Up: func(db *gorm.DB) error {
			err := db.AutoMigrate(&migtest{})
			return err
		},
		Down: func(db *gorm.DB) error {
			err := db.Migrator().DropTable("migtests")
			return err
		},
	})
	Mig(State{
		Level: "mig002",
		Tag:   "add_entries",
		Up: func(db *gorm.DB) error {
			err := db.Create(&migtest{Name: "Tester", Email: "test@test.tx"}).Error
			return err
		},
		Down: func(db *gorm.DB) error {
			err := db.Delete(migtest{}, "Name = 'Tester'").Error
			return err
		},
	})
	Mig(State{
		Level: "mig003",
		Tag:   "add_column",
		Up: func(db *gorm.DB) error {
			type migtest struct {
				Testcol string `gorm:"size:200"`
			}
			err := db.Migrator().AddColumn(&migtest{}, "testcol")
			return err
		},
		Down: func(db *gorm.DB) error {
			err := db.Migrator().DropColumn(&migtest{}, "testcol")
			return err
		},
	})
}

func Test_performMigration(t *testing.T) {

	createMockMigrationList()

	db, err := gorm.Open(postgres.Open("host=localhost user=user password=passpass dbname=testdb port=5432 sslmode=disable"), &gorm.Config{Logger: logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel: logger.Silent, // Log level
		},
	)})
	t.Run("DB ok", func(t *testing.T) {
		assert.NoError(t, err)
	})

	// Pre clean up
	err = db.Migrator().DropTable("migrations")
	t.Run("Migrator 1", func(t *testing.T) {
		assert.NoError(t, err)
	})
	db.Migrator().DropTable("migtests")
	t.Run("Migrator 2", func(t *testing.T) {
		assert.NoError(t, err)
	})

	migrationStore := NewMigrationStore(db)

	type args struct {
		fromTag string
		toTag   string
		user    string
	}
	tests := []struct {
		name       string
		args       args
		currentTag string // Current tag AFTER! Migration is executed
		wantErr    bool
	}{
		{
			name:    "nulltest",
			wantErr: false,
			args: args{
				fromTag: "null",
				toTag:   "null",
				user:    "testuser",
			},
			currentTag: "",
		},
		{
			name:    "no user name",
			wantErr: true,
			args: args{
				fromTag: "null",
				toTag:   "add_entries",
				user:    "",
			},
			currentTag: "",
		},
		{
			name:    "wrong start to UP mig003",
			wantErr: true,
			args: args{
				fromTag: "add_entries",
				toTag:   "add_column",
				user:    "fooname",
			},
			currentTag: "",
		},
		{
			name:    "UP to mig002 two steps",
			wantErr: false,
			args: args{
				fromTag: "null",
				toTag:   "add_entries",
				user:    "testuser",
			},
			currentTag: "add_entries",
		},
		{
			name:    "repeat UP mig002",
			wantErr: true,
			args: args{
				fromTag: "null",
				toTag:   "add_entries",
				user:    "testuser",
			},
			currentTag: "add_entries", // stays the same as before
		},
		{
			name:    "same stage no changes",
			wantErr: false,
			args: args{
				fromTag: "add_entries",
				toTag:   "add_entries",
				user:    "testuser",
			},
			currentTag: "add_entries",
		},
		{
			name:    "UP mig003 one step",
			wantErr: false,
			args: args{
				fromTag: "add_entries",
				toTag:   "add_column",
				user:    "fooname",
			},
			currentTag: "add_column",
		},
		{
			name:    "repeat UP mig003",
			wantErr: true,
			args: args{
				fromTag: "add_entries",
				toTag:   "add_column",
				user:    "fooname",
			},
			currentTag: "add_column",
		},
		{
			name:    "DOWN mig002",
			wantErr: false,
			args: args{
				fromTag: "add_column",
				toTag:   "add_entries",
				user:    "fooname",
			},
			currentTag: "add_entries",
		},
		{
			name:    "repeat DOWN mig002",
			wantErr: true,
			args: args{
				fromTag: "add_column",
				toTag:   "add_entries",
				user:    "fooname",
			},
			currentTag: "add_entries",
		},
		{
			name:    "DOWN to null", // more than two steps down is not allowed
			wantErr: true,
			args: args{
				fromTag: "add_entries",
				toTag:   "null",
				user:    "fooname",
			},
			currentTag: "add_entries",
		},
		{
			name:    "DOWN to mig001",
			wantErr: false,
			args: args{
				fromTag: "add_entries",
				toTag:   "create_migtest",
				user:    "fooname",
			},
			currentTag: "create_migtest",
		},
		{
			name:    "DOWN to null",
			wantErr: false,
			args: args{
				fromTag: "create_migtest",
				toTag:   "null",
				user:    "fooname",
			},
			currentTag: "null",
		},
		{
			name:    "UP to mig001", // again, one step up
			wantErr: false,
			args: args{
				fromTag: "null",
				toTag:   "create_migtest",
				user:    "bar name",
			},
			currentTag: "create_migtest",
		},
		{
			name:    "DOWN to null", // again, one step down to null state
			wantErr: false,
			args: args{
				fromTag: "create_migtest",
				toTag:   "null",
				user:    "fooname",
			},
			currentTag: "null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := performMigration(tt.args.fromTag, tt.args.toTag, tt.args.user, db, migrationStore)

			currentTag, errCurrenTag := migrationStore.GetCurrentTag()
			assert.Equal(t, tt.currentTag, currentTag)
			if currentTag == "" {
				assert.Error(t, errCurrenTag)
			} else {
				assert.NoError(t, errCurrenTag)
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}

	// After clean up
	err = db.Migrator().DropTable("migrations")
	t.Run("Migrator 1", func(t *testing.T) {
		assert.NoError(t, err)
	})
	db.Migrator().DropTable("migtest")
	t.Run("Migrator 2", func(t *testing.T) {
		assert.NoError(t, err)
	})
}
