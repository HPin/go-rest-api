// main.go

package main

import "os"

// entry point for the application
func main() {
	// init app
	a := App{}

	// init db with credentials stored in environment variables
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	// run the app on port 8010
	a.Run(":8010")
}
