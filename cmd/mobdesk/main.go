package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ericklucioh/mobdesk/internal/cobra"
)

func main() {
	if err := cobra.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, cobra.ErrStatusStrict) {
			os.Exit(3)
		}
		os.Exit(1)
	}
}
