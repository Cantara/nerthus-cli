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
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	log "github.com/cantara/bragi/sbragi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Server struct {
	Host  string   `json:"host"`
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

var serversCache *[]Server

var (
	zeroDialer net.Dialer
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func init() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return zeroDialer.DialContext(ctx, "tcp4", addr)
	}
	httpClient.Transport = transport
}

func getServers(profile Profile) (servers []Server) {
	if serversCache != nil {
		return *serversCache
	}
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/servers", profile.nerthusHost),
		nil,
	)
	if err != nil {
		log.WithError(err).Error("while creating request")
		panic(err)
	}
	req.SetBasicAuth(profile.username, profile.password)
	// tr := &http.Transport{
	// 	TLSHandshakeTimeout: 30 * time.Second,
	// 	DisableKeepAlives:   true,
	// }
	// client := &http.Client{Transport: tr}
	r, err := httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("while executing request")
		panic(err)
	}
	err = json.NewDecoder(r.Body).Decode(&servers)
	if err != nil {
		log.WithError(err).Error("while reading request")
		panic(err)
	}
	serversCache = &servers
	return
}

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Command to ssh onto a node",
	Long: `Helps ssh onto a single node. 
Can select what user and what node to ssh to`,
	Args: cobra.RangeArgs(2, 3), // cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		profile := GetProfile(args)
		servers := getServers(profile)
		serverNames := make([]string, len(servers))
		i := 0
		serverInfo := map[string]Server{}
		for _, server := range servers {
			serverNames[i] = server.Name
			serverInfo[server.Name] = server
			i++
		}
		hostname := args[1]
		hostInfo, ok := serverInfo[hostname]
		if !ok {
			log.Fatal("host is not found", "hostname", hostname, "serverNames", serverNames)
		}
		switch len(args) {
		case 2:
			ssh("ec2-user", hostInfo.Host)
		case 3:
			ssh(args[2], hostInfo.Host)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		initConfig()
		profile := GetProfile(args)
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
		servers := getServers(profile)
		serverNames := make([]string, len(servers))
		i := 0
		serverInfo := map[string]Server{}
		for _, server := range servers {
			serverNames[i] = server.Name
			serverInfo[server.Name] = server
			i++
		}

		if len(args) == 1 {
			return serverNames, cobra.ShellCompDirectiveNoFileComp
		}

		hostName := args[1]
		return serverInfo[hostName].Users, cobra.ShellCompDirectiveNoFileComp
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

func ssh(user, host string) {
	fmt.Printf("ssh %s@%s\n", user, host)
	binary, lookErr := exec.LookPath("ssh")
	if lookErr != nil {
		panic(lookErr)
	}
	syscall.Exec(binary, []string{"ssh", fmt.Sprintf("%s@%s", user, host)}, os.Environ())
}
