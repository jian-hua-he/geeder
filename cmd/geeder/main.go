package main

import (
	"log"
	"os"

	"github.com/jian-hua-he/geeder/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
