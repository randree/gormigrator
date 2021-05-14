package gormigrator

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"

	"gorm.io/gorm"
)

// To collect all up and down functions and a security code we use
type migrationFileListMap map[string]*updown

var migrationFileList = make(migrationFileListMap)

// Code must be called first in init followed by Up and Down
func Code(code string) {
	filename := caller()
	codeExists(code, filename)
	if code == "" {
		log.Fatal("code is missing (" + filename + ")")
	}
	if code == "null" {
		log.Fatal("code can't have reserved name null (" + filename + ")")
	}
	if strings.Contains(code, " ") {
		log.Fatal("code contains a whitespace (" + filename + ")")
	}
	migrationFileList[filename] = &updown{
		level: filename,
		code:  code,
	}
}

// Up must be called from init
func Up(upgrader func(*gorm.DB) error) {
	filename := caller()
	current := migrationFileList[filename]
	if current == nil {
		log.Fatal("a code is missing at " + filename)
	}
	current.up = upgrader
}

// Down must be called from init
func Down(downgrader func(*gorm.DB) error) {
	filename := caller()
	current := migrationFileList[filename]
	if current == nil {
		log.Fatal("a code is missing at " + filename)
	}
	current.down = downgrader
}

func consistencyFileCheck(list *migrationFileListMap) {
	for name, entry := range *list {
		if entry.up == nil {
			log.Fatal("a upgrader is missing at " + name)
		}
		if entry.down == nil {
			log.Fatal("a downgrader is missing at " + name)
		}
	}
}

func caller() string {
	_, path, _, _ := runtime.Caller(2)
	return filepath.Base(path)
}

func codeExists(newCode string, filename string) {
	for _, updownEntry := range migrationFileList {
		if updownEntry.code == newCode {
			log.Fatal("code already exists in " + filename)
		}
	}
}
