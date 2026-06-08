package main

import (
	"context"
	"os"

	"github.com/opensoha/soha-cli/internal/sohacli"
)

func main() {
	os.Exit(sohacli.Run(context.Background(), os.Args[1:], sohacli.Runtime{}))
}
