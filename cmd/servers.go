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
	"slices"
	"strings"
	"time"

	"github.com/acarl005/textcol"
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
		return zeroDialer.DialContext(ctx, "tcp4", addr) //Should be able to remove ipv4 block soon
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
	//TODO HANDLE HTTP STATUS AND PRINT ERROR MESSAGE
	err = json.NewDecoder(r.Body).Decode(&servers)
	if err != nil {
		log.WithError(err).Error("while reading request")
		panic(err)
	}
	serversCache = &servers
	return
}

var serversCmd = &cobra.Command{
	Use:   "servers <profile>",
	Short: "Lists all servers in profile",
	Long:  `Lists all servers in a profile`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		profile := GetProfile(args)
		serversInfo := getServers(profile)
		servers := make([]string, len(serversInfo))
		for i, server := range serversInfo {
			servers[i] = server.Name
		}
		slices.Sort[[]string, string](servers)
		textcol.PrintColumns(&servers, 4)
		/*
			servers := map[string][]string{}
			for _, server := range serversInfo {
				env := strings.Split(server.Name, "-")[0]
				servers[env] = append(servers[env], server.Name)
			}
			for k, v := range servers {
				fmt.Println(k)
				textcol.PrintColumns(&v, 4)
			}
		*/
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
	rootCmd.AddCommand(serversCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
