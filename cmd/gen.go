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
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/cantara/bragi/sbragi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return zeroDialer.DialContext(ctx, "tcp4", addr) //Should be able to remove ipv4 block soon
	}
	httpClient.Transport = transport
}

var genCmd = &cobra.Command{
	Use:   "gen [profile]",
	Short: "Generates scripts",
	Long: `Generates a set of scripts to help use nerthus cli.
Used for createing jumphosts or decrease the distance from one user to another.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var profileNames []string
		if len(args) == 1 {
			profileNames = append(profileNames, args[0])
		} else {
			for _, profileName := range viper.AllKeys() {
				if !strings.HasSuffix(profileName, ".nerthus") {
					continue
				}
				profileNames = append(profileNames, strings.TrimSuffix(profileName, ".nerthus"))
			}
		}
		sshdir := "ssh"
		os.Mkdir(sshdir, 0750)
		pdir := "prov"
		os.Mkdir(pdir, 0750)
		for i := range profileNames {
			profile := GetProfile(profileNames[i : i+1])
			func(dir string) {
				os.Mkdir(dir, 0750)
				serversInfo := getServers(profile)
				for _, server := range serversInfo {
					path := filepath.Clean(fmt.Sprintf("%s/ssh_%s.sh", dir, server.Name))
					f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0750)
					log.WithError(err).Trace("opening ssh file")
					if err != nil {
						continue
					}
					defer func() {
						log.WithError(f.Close()).Trace("closing file", "path", path)
					}()
					fmt.Fprintf(f, "#This script is managed by nerthus-cli, do not edit!\nnerthus-cli ssh %s %s\n", profile.name, server.Name)

				}
				for _, env := range getEnvs(profile) {
					if env == "" {
						continue
					}
					path := filepath.Clean(fmt.Sprintf("%s/%s-%s.sh", pdir, profile.name, env))
					f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0750)
					log.WithError(err).Trace("opening ssh file")
					if err != nil {
						return
					}
					defer func() {
						log.WithError(f.Close()).Trace("closing file", "path", path)
					}()
					fmt.Fprintf(f, "#This script is managed by nerthus-cli, do not edit!\nnerthus-cli conf %s %s\n", profile.name, env)
				}
			}(filepath.Clean(fmt.Sprintf("%s/%s", sshdir, profile.name)))
		}
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
	rootCmd.AddCommand(genCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
