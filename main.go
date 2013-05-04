package main

import (
	"flag"
	"fmt"
	"rsswrapper/rssw"
)

var ip = flag.Bool("d", false, "Save to database")

func main() {
	flag.Parse()
	if *ip {
		fmt.Println("Database writing mode")
		rssw.WriteToDatabase()
	} else {
		rssw.Start()
	}
}
