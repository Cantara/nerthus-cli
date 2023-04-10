/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// certCmd represents the cert command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Key will manage your public ssh key",
	Long: `Key will manage your public ssh key
Key can: 
Add a new pub key to Nerthus
List the pub keys you have added to Nerthus
Delete pub key from Nerthus`,
}

func init() {
	rootCmd.AddCommand(keyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// certCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// certCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
