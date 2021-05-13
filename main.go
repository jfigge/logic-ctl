package main

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/cmd"
	"log"
)

// go:generate swagger generate spec -o ./swaggerui/swagger-spec.json --scan-models --exclude-deps
func main() {
	log.Fatal(cmd.Execute())
}
