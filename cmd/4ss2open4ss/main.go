package main

import (
	"fmt"
	"os"

	"github.com/Dangthrimble/4ss24ssxml/internal/convert"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: 4ss2open4ss <file.4ss>\n")
		os.Exit(2)
	}
	input := os.Args[1]
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	xml, err := convert.Convert(data, convert.SetlistTitle(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "convert: %v\n", err)
		os.Exit(1)
	}
	out := convert.OutputPath(input)
	if err := os.WriteFile(out, xml, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(out)
}
