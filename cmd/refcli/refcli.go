package main

import (
	"fmt"
	"os"

	"github.com/loveyourstack/lys-ref/cmd/refcli/rootcli"
)

func main() {
	if err := rootcli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
