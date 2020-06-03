// main.go

package main

// entry point for the application
func main() {
	// init app
	a := App{}

	// init db with credentials stored in environment variables
	a.Initialize(
		"postgres",
		"postgres",
		"postgres")

	// run the app on port 8010
	a.Run(":8010")
}
