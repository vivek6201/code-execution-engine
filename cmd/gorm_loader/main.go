package main

import (
	"fmt"
	"io"
	"os"

	"github.com/code-execution-engine/internal/judge"
	"github.com/code-execution-engine/internal/models"
	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&models.User{},
		&models.APIKey{},
		&judge.JobRecord{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
