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
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return zeroDialer.DialContext(ctx, "tcp4", addr)
	}
	httpClient.Transport = transport
}

func getEnvs(profile Profile) (envs []string) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("https://%s/env", profile.nerthusHost),
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
	err = json.NewDecoder(r.Body).Decode(&envs)
	if err != nil {
		log.WithError(err).Error("while reading request")
		panic(err)
	}
	return
}

func execConfig(profile Profile, path string) (envs []string) {
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("https://%s/config/%s", profile.nerthusHost, path),
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
	_, err = httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("while executing request")
		panic(err)
	}
	/*
		//TODO HANDLE HTTP STATUS AND PRINT ERROR MESSAGE
		err = json.NewDecoder(r.Body).Decode(&envs)
		if err != nil {
			log.WithError(err).Error("while reading request")
			panic(err)
		}
	*/
	return
}

// sshCmd represents the ssh command
var confCmd = &cobra.Command{
	Use:   "conf <profile> <env> [system] [cluster] [service]",
	Short: "Command to execute config deployment",
	Long: `Helps deploy config changes.
Can execute on every level, from the whole solution to a single service`,
	Args: cobra.RangeArgs(2, 5), // cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		profile := GetProfile(args)
		envs := getEnvs(profile)
		env := args[1]
		i := ArrayContains(envs, env)
		if i < 0 {
			log.Fatal("host is not found", "env", env, "envs", envs)
		}

		execConfig(profile, strings.Join(args[1:], "/"))
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
		if len(args) > 2 { //TODO: add completion for the remaining values
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		envs := getEnvs(profile)

		return envs, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(confCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ArrayContains(arr []string, val string) int {
	for i, v := range arr {
		if v == val {
			return i
		}
	}
	return -1
}
