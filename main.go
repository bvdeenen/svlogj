/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"io/fs"
	"os"
	"svlogj/cmd"

	"github.com/spf13/cobra"
)

func main() {

	file, err := os.Open("/var/log/socklog")
	if err != nil {
		switch err.(type) {
		case *fs.PathError:
			fmt.Print(`Permissions error!

We need to be able to read recursively from /var/log/socklog,
for the 'config' files, as well as for reading the log via 'svlogtail'.

Add yourself to the socklog group

	sudo usermod -aG socklog $USER

and then log out and log in again (or use 'newgrp socklog' in an existing shell)
`)
			os.Exit(1)
			break
		default:
			cobra.CheckErr(err)
		}
	}
	err = file.Close()
	cobra.CheckErr(err)
	cmd.Execute()
}
