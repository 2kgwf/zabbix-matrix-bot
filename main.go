package main

import (
	"log"
	"os"
	"strings"
	"zabbix-matrix-bot/bot"
)

func main() {
	homeserverURL := ""
	userID := ""
	accessToken := ""
	zabbixAPIURL := ""
	zabbixUsername := ""
	zabbixPassword := ""
	admin := ""

	for _, e := range os.Environ() {
		split := strings.SplitN(e, "=", 2)
		switch split[0] {
		case "ZABBIX_MATRIX_HOMESERVER_URL":
			homeserverURL = split[1]
		case "ZABBIX_MATRIX_USER_ID":
			userID = split[1]
		case "ZABBIX_MATRIX_ACCESS_TOKEN":
			accessToken = split[1]
		case "ZABBIX_API_URL":
			zabbixAPIURL = split[1]
		case "ZABBIX_USERNAME":
			zabbixUsername = split[1]
		case "ZABBIX_PASSWORD":
			zabbixPassword = split[1]
		case "ZABBIX_MATRIX_ADMIN":
			accessToken = split[1]
		}
	}

	if len(os.Args) > 7 {
		homeserverURL = os.Args[1]
		userID = os.Args[2]
		accessToken = os.Args[3]
		zabbixAPIURL = os.Args[4]
		zabbixUsername = os.Args[5]
		zabbixPassword = os.Args[6]
		admin = os.Args[7]
	}

	if homeserverURL == "" || userID == "" || accessToken == "" || zabbixAPIURL == "" || zabbixUsername == "" || zabbixPassword == "" || admin == "" {
		log.Fatal("invalid config")
	}

	log.Fatal(bot.NewZabbixBot(homeserverURL, userID, accessToken, zabbixAPIURL, zabbixUsername, zabbixPassword, admin).Run())
}
