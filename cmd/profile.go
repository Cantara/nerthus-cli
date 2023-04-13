package cmd

import (
	log "github.com/cantara/bragi/sbragi"
	"github.com/spf13/viper"
)

type Profile struct {
	name        string
	username    string
	password    string
	nerthusHost string
}

func GetDefault(args []string) (profile string) {
	profile = viper.GetString("profile")
	if len(args) == 1 {
		if !viper.IsSet(args[0]) {
			log.Fatal(
				"provided profile was not set",
				"arg",
				args[0],
			)
			return
		}
		profile = args[0]
	}
	return
}

func GetProfile(args []string) (profile Profile) {
	profileName := GetDefault(args)
	profileMap := viper.GetStringMapString(profileName)
	profile = Profile{
		name:        profileName,
		username:    profileMap["username"],
		password:    profileMap["password"],
		nerthusHost: profileMap["nerthus"],
	}
	return
}
