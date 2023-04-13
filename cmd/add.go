/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var homeDir, _ = os.UserHomeDir()

type key struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func putKey(cmd *cobra.Command, profile Profile) (servers []Server) {
	path, _ := cmd.Flags().GetString("path")
	name, _ := cmd.Flags().GetString("name")
	owner, _ := cmd.Flags().GetString("owner")
	pubKey, err := os.ReadFile(path)
	if err != nil {
		log.WithError(err).Error("while reading request data")
		panic(err)
	}
	k := key{
		Name: name,
		Data: string(pubKey),
	}
	body, err := jsoniter.Marshal(k)
	if err != nil {
		log.WithError(err).Error("while marshaling request")
		panic(err)
	}
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("https://%s/key/%s/%s", profile.nerthusHost, owner, name),
		bytes.NewReader(body),
	)
	if err != nil {
		log.WithError(err).Error("while creating request")
		panic(err)
	}
	req.SetBasicAuth(profile.username, profile.password)
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.WithError(err).Error("while executing request")
		panic(err)
	}
	return
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new public ssh cert to Nerthus",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		putKey(cmd, GetProfile(args))
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		initConfig()
		if len(args) == 0 {
			var profileNames []string
			for _, profileName := range viper.AllKeys() {
				if !strings.HasSuffix(profileName, ".nerthus") {
					continue
				}
				profileNames = append(profileNames, strings.TrimSuffix(profileName, ".nerthus"))
			}
			return profileNames, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	keyCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	hostName, _ := os.Hostname()
	addCmd.Flags().StringP("path", "p", homeDir+"/.ssh/id_rsa.pub", "Path for public ssh cert")
	addCmd.Flags().StringP("name", "n", hostName, "Friendly name of cert")
	addCmd.Flags().StringP("owner", "o", "", "Owner of cert")
	addCmd.MarkFlagRequired("owner")
}
