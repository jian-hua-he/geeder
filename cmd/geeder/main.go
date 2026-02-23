package main

import (
	"embed"

	"github.com/jian-hua-he/geeder"
)

//go:embed seeds/*.sql
var seeds embed.FS

func main() {
	geeder.Main(seeds)
}
