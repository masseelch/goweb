package main

import (
	"github.com/masseelch/gowebapp"
)

func main() {
	fs := gowebapp.ParseFlags()

	gowebapp.GenerateRepositories(fs.Repository)
}
