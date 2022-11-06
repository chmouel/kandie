package main

import (
	"fmt"
	"os"

	kandie "github.com/chmouel/kandie/pkg"
)

func main() {
	if err := kandie.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stdout, "kandie %s\n", err.Error())
	}
}
