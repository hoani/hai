package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	tuiinit "github.com/hoani/hai/tui/init"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the hai assistant.",
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(tuiinit.New())
		_, err := p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
