# GORMigrator - A migration tool based on GORM

The GORMigrator is a lightweight but powerful and flexible migration tool based on GORM. Especially useful in container environments like Docker.

The goal is to create a build of your migration setup. You can put this build in a `FROM scratch AS bin` container of minimal size and deploy it to your server where you can perform the migration.

## Example

Start with creating a migration directory. 
```bash
$ mkdir migration
```

Create a `main.go` in this directory. It is needed to define the database connection.

```golang
package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/caarlos0/env/v6"
	"github.com/randree/gormigrator"
)

type DBconfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	DBname   string `env:"DB_PORT" envDefault:"testdb"`
	User     string `env:"DB_USER" envDefault:"user"`
	Password string `env:"DB_PASSWORD" envDefault:"passpass"`
}

func main() {
	cfg := DBconfig{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.DBname, cfg.Port)

	// Or choose other dialects like MySQL
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		fmt.Println(err.Error())
	}

	gormigrator.InitMigration(db)
}
```



Next, we create the first migration file `mig0001.go`:
```golang
package main

import (
	g "github.com/randree/gormigrator"
	"gorm.io/gorm"
)

func init() {
	g.Code("first_migration")
	g.Up(func(db *gorm.DB) error {

		type migtest struct {
			gorm.Model
			Name  string `gorm:"size:255"`
			Email string `gorm:"size:300"`
		}

		err := db.AutoMigrate(&migtest{})

		return err
	})

	g.Down(func(db *gorm.DB) error {

		err := db.Migrator().DropTable("migtests")

		return err
	})

}
```
All migration files must to be ordered.

Lets create the next migration `mig0002.go`:
```golang
package main

import (
	g "github.com/randree/gormigrator"
	"gorm.io/gorm"
)

func init() {
	type migtest struct {
		// we need only new column in struct
		Testcol string
	}
	g.Code("adding_new_column")
	g.Up(func(db *gorm.DB) error {

		err := db.Migrator().AddColumn(&migtest{}, "testcol")

		return err
	})

	g.Down(func(db *gorm.DB) error {
		err := db.Migrator().DropColumn(&migtest{}, "testcol")
		return err
	})

}
```

Now we are going to create a module.

```console
$ go mod init migration
```

The migration to the first level upgrade is done by
```console
$ go run ./... -from null -to first_migration -user myname
```
`null` represents the zeroth level. `first_migration` is the code we defined in `mig0001.go`.

For the second upgrade we use
```console
$ go run ./... -from first_migration -to adding_new_column -user myname
```
So we upgraded from the first stage `first_migration` to `adding_new_column` which is the code from the second migration file `mig0002.go`.

We also could have used 
```console
$ go run ./... -from null -to adding_new_column -user myname
```
to upgrade from level 0 to 2.

Downgrading is similar:
```console
$ go run ./... -from adding_new_column -to first_migration -user myname
```
and than back to `null` (level 0) with
```console
$ go run ./... -from first_migration -to null -user myname
```
You can downgrade only one step at a time.


## How it works

For migrating UP or DOWN we are using a code instead of `up` or `down` or filenames. It helps you to be aware of what the migration does. `mig0001` and `mig0002` do not give you any information about the process, contrary to `first_migration` and `adding_new_column`. Moreover, after compilation, these codes are "hidden" from others who do not know the source code. This provides an additional layer of security.

Example of up and downgrades:

| Upgrades | Downgrades | Level |
|-------|-------|---|
| adding_new_column | adding_new_column | 2 |
| ⬆ `-from first_migration -to adding_new_column` | ⬇ `-from adding_new_column -to first_migration` | |
| first_migration | first_migration | 1 |
| ⬆ `-from null -to first_migration` | ⬇ `-from first_migration -to null` | |
| null | null | 0 |

The first migration starts with `null`, which represents an empty database. The files should be in the correct order. For example, you can choose `mig0001.go`, `test01.go` or `foo0000001.go`.

### Flags

You can use the following flags:
| Flags | Description |
|-----|---|
| `-list` | List all migrations |
| `-version` | Show version |
| `-from <CODE>` | From-code |
| `-to <CODE>` | To-code |
| `-user <NAME>` | Admin or username |

### Username

Use `-user <NAME>` to document the user performing the migration.

### List

`-list` gives you something like that:
```
| DATETIME HISTORY                         | LEVEL (State)                            | USER                 |
| ---------------------------------------- | ---------------------------------------- | -------------------- |
| 2021-05-14 16:03:31.393752 +0200 CEST    | mig0003.go                     (current) | Dan                  |
| 2021-05-11 16:03:31.351858 +0200 CEST    | mig0002.go                               | Maddy                |
| 2021-05-02 16:03:31.310811 +0200 CEST    | mig0001.go                               | The Downgrader       |
| 2021-04-24 09:03:31.265472 +0200 CEST    | mig0002.go                               | Team Migration       |
| 2021-03-14 11:03:31.228521 +0200 CEST    | mig0001.go                               | Maddy                |
| 2021-03-12 09:03:31.204998 +0200 CEST    | null                                     | Dan                  |
```