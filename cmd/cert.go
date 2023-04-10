/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// certCmd represents the cert command
var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Cert will manage your public ssh cert",
	Long: `Cert will manage your public ssh cert
Cert can: 
Add a new cert to Nerthus
List the cert you have added to Nerthus
Delete certs from Nerthus`,
}

func init() {
	rootCmd.AddCommand(certCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// certCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// certCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
