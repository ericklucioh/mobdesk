package main

import (
	"fmt"
	"github.com/ericklucioh/mobdesk/internal/cobra"
	"os"
)

func main() {
	if err := cobra.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
