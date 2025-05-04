package main

import (
	"gping/src/data/show"

	"github.com/gookit/color"
)

func main() {
	pw := show.NewWindowProgram()
	if _, err := pw.Run(); err != nil {
		color.Redln(err)
	}
}