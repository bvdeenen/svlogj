/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"svlogj/cmd"
	"svlogj/pkg/utils"

	"github.com/spf13/cobra"
)

func main() {

	file, err := os.Open(utils.SocklogDir())
	if err != nil {
		var pathError *fs.PathError
		switch {
		case errors.As(err, &pathError):
			fmt.Printf(`Permissions error!

We need to be able to read recursively from %s,
for the 'config' files, as well as for reading the log via 'svlogtail'.

Add yourself to the socklog group

	sudo usermod -aG socklog $USER

and then log out and log in again (or use 'newgrp socklog' in an existing shell)
`, utils.SocklogDir())
			os.Exit(1)
		default:
			cobra.CheckErr(err)
		}
	}
	err = file.Close()
	cobra.CheckErr(err)
	cmd.Execute()
}
