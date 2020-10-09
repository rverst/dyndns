package main

import "github.com/integrii/flaggy"

var (
	version = "unknown"
)

func main() {
	flaggy.SetName("dyndns")
	flaggy.SetDescription("Server for dynamic DNS updates.")
	flaggy.SetVersion(version)


}
