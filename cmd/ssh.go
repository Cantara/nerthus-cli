/*
Copyright © 2023 Sindre Brurberg sindre@brurberg.no

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type Server struct {
	Host  string   `json:"host"`
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

func getServers() (servers []Server) {
	r, err := http.Get("https://nerthus.greps.dev/servers")
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(r.Body).Decode(&servers)
	if err != nil {
		panic(err)
	}
	return
}

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Command to ssh onto a node",
	Long: `Helps ssh onto a single node. 
Can select what user and what node to ssh to`,
	Args: cobra.RangeArgs(1, 2), // cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 1:
			fmt.Printf("ssh %s@%s\n", "ec2-user", args[0])
		case 2:
			fmt.Printf("ssh %s@%s\n", args[1], args[0])
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		servers := getServers()
		serverNames := make([]string, len(servers))
		i := 0
		serverInfo := map[string]Server{}
		for _, server := range servers {
			serverNames[i] = server.Name
			serverInfo[server.Name] = server
			i++
		}

		switch len(args) {
		case 0:
			return serverNames, cobra.ShellCompDirectiveNoFileComp
		case 1:
			return serverInfo[args[0]].Users, cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
