package main

import (
	"github.com/ericklucioh/mobdesk/internal/cobra"
	"os"
	"fmt"
)

func main() {
	if err := cobra.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}