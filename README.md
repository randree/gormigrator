# GORMigrator v1 - A migration tool based on GORM

The GORMigrator is a lightweight but powerful and flexible migration tool based on GORM. Especially useful in container environments like Docker.

The goal is to create a build of your migration setup.

## Steps for deployment

* Create a go build
* Putting build in a `FROM scratch AS bin` container for minimal size
* Deploy
* Start service or container with environment variables (`FROM=null TO=create_user_table USER=testuser ...` to perform migration on database

## Example

See example under `example/` to see how it works.

The folder looks like 
```bash
├── migration
│   ├── main.go
│   ├── mig00001.go
│   ├── mig00002.go
│   └── mig00003.go
    ...
```

`mig`-files can have any names, as long as they are ordered (e.g. `mig000xx.go`).
Example for `main.go`:
```golang
package main

import (
	"fmt"

	g "github.com/randree/gormigrator/v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.Open("host=localhost user=user password=passpass dbname=testdb port=5432 sslmode=disable"))
	if err != nil {
		fmt.Println(err.Error())
	}

	g.InitMigration(db)
}
```

Mig-files (e.g. `mig001.go`) looks like:
```golang
package main

import (
	g "github.com/randree/gormigrator/v1"
	"gorm.io/gorm"
)

func init() {

	// Mig function to load state 
	g.Mig(g.State{

		// Tag: Name for state after migration
		Tag: "roles",

		// Up-function to migrate up
		Up: func(db *gorm.DB) error {

			type Role struct {
				ID   int `gorm:"primarykey"`
				Role string
			}
			db.AutoMigrate(&Role{})
            err := db.Create(&Role{
				ID:   2,
				Role: "Guest",
			}).Error

			return err
		},

		// Down-function to migrate down
		Down: func(db *gorm.DB) error {
			err := db.Migrator().DropTable("roles")
			return err
		},
	})
}
```

## Create and use module

Steps to create a go module:
```console
$ go mod init migration
```
To load dependencies:
```console
$ go mod tidy
```

To run a migration:
```console
$ FROM=null TO=users USER=foo go run ./...

⬆ UPGRADE (0) ⟶ (mig00001.go) tag: null ⟶ roles 
⬆ UPGRADE (mig00001.go) ⟶ (mig00002.go) tag: roles ⟶ customers 
⬆ UPGRADE (mig00002.go) ⟶ (mig00003.go) tag: customers ⟶ users
```
To initialize first tables you start migration from a tag with name `null`.

To show a migration history:
```console
$ HISTORY=1 go run ./...

| DATETIME HISTORY                       | LEVEL (State)         | USER       |
| -------------------------------------- | --------------------- | ---------- |
| 2021-08-31 13:09:47.932619 +0200 CEST  | mig00003.go (current) | foo        |
| 2021-08-31 13:09:47.908876 +0200 CEST  | mig00002.go           | foo        |
| 2021-08-31 13:09:47.871357 +0200 CEST  | mig00001.go           | foo        |
```

To show version:
```console
$ VERSION=1 go run ./...

Gormigrator version:  1.0.0
```

Or combine all:
```console
$ VERSION=1 HISTORY=1 FROM=users TO=testers USER=foo go run ./...
```
(`HISTORY` gives you the the list of migrations before any action)

### Calls

```console
$ FROM=<Tag> TO=<Tag> [HISTORY=1] [VERSION=1] (Docker Container | go build | go run ./...)
```
Docker-compose file
```yaml
...
  migrator:
    image: from_scratch_image
    environment:
      FROM: <Tag>
      TO: <Tag>
...

```

## Downgrade note

To prevent a accidental and fatal downgrade to `null` or too many steps in a row only one downgrade step at a time is allowed.

## References

- [GROM](https://gorm.io/) The GORM project.
- [GROM migration methods.](https://gorm.io/docs/migration.html) You can use these methods in the GORMigrator.