package gormigrator

import "log"

// In case of errorLog we don't want to inject a logger to the migrator
// The reason is to keep it as small as possible later in main
func errorLog(message string) {
	if Testing {
		// can be caught by testing
		panic(message)
	} else {
		log.Fatal(message)
	}
}
