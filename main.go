package main

import (
	"github.com/kumparan/machinerydash/console"
	"github.com/markbates/pkger"
)

func main() {
	// include all files to be bundled at build times
	pkger.Include("/views")
	console.Execute()
}
