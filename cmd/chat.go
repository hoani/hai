/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	tuichat "github.com/hoani/hai/tui/chat"
	tuiinit "github.com/hoani/hai/tui/init"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		chat, err := tuichat.New()
		if err != nil { // Attempt to init.
			p := tea.NewProgram(tuiinit.New())
			if _, err := p.Run(); err != nil {
				return err
			}

			chat, err = tuichat.New()
			if err != nil {
				return err
			}
		}
		p := tea.NewProgram(chat)
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
