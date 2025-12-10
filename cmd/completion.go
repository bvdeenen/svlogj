// Package cmd 
package cmd

import (
	"svlogj/pkg/utils"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "generate a completion for your shell",
	Long: `svlogj completion bash|zsh|fish --help to see tips for installing the completion`,
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.AddCommand(bashCompletion)
	completionCmd.AddCommand(zshCompletion)
	completionCmd.AddCommand(fishCompletion)
}

var bashCompletion = &cobra.Command{
	Use:                        "bash",
	Long:                       `Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

        source <(svlogj completion bash)

To load completions for every new session, execute once:

#### Linux:

        svlogj completion bash > ~/.bash_completion.d/svlogj or
        svlogj completion bash > /etc/bash_completion.d/svlogj

#### macOS:

        svlogj completion bash > /usr/local/etc/bash_completion.d/svlogj

You will need to start a new shell for this setup to take effect.
	`,
	ValidArgsFunction:          utils.NoFilesEmptyCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
	},
}
var zshCompletion = &cobra.Command{
	Use:                        "zsh",
	Long: `
	Generate the autocompletion script for the zsh shell.

	If shell completion is not already enabled in your environment you will need
	to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

	To load completions for every new session, execute once:

	#### Linux:

	svlogj completion zsh > "${fpath[1]}/_svlogj"

	#### macOS:

	svlogj completion zsh > /usr/local/share/zsh/site-functions/_svlogj

	You will need to start a new shell for this setup to take effect.`,
	ValidArgsFunction:          utils.NoFilesEmptyCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
	},
}
var fishCompletion = &cobra.Command{
	Use:                        "fish",
	Long: `
Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

        svlogj completion fish | source

To load completions for every new session, execute once:

        svlogj completion fish > ~/.config/fish/completions/svlogj.fish

You will need to start a new shell for this setup to take effect.
	`,
	ValidArgsFunction:          utils.NoFilesEmptyCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
	},
}
