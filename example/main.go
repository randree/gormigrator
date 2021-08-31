package main

import (
	"fmt"

	"github.com/randree/gormigrator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.Open("host=localhost user=user password=passpass dbname=testdb port=5432 sslmode=disable"))
	if err != nil {
		fmt.Println(err.Error())
	}

	gormigrator.InitMigration(db)
}
