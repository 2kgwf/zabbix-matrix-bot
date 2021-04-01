package bot

import (
	"log"
	"strings"
	"zabbix-matrix-bot/matrix"

	"github.com/matrix-org/gomatrix"
)

// ZabbixBot contains all of the actual bot logic
type ZabbixBot struct {
	client                        matrix.Client
	zabbixIdentifierToMatrixEvent map[string]string
	zabbixAPIURL                  string
	zabbixUsername                string
	zabbixPassword                string
	adminUser                     string
}

func (bot ZabbixBot) handleMemberEvent(event *gomatrix.Event) {
	if event.Content["membership"] == "invite" && *event.StateKey == bot.client.UserID {
		if event.Sender == bot.adminUser {
			bot.client.JoinRoom(event.RoomID)
			log.Print("Joined room " + event.RoomID)
		} else {
			log.Print("Ignoring room invite " + event.RoomID)
		}
	}
}

func (bot ZabbixBot) handleTextEvent(event *gomatrix.Event) {
	if event.Content["msgtype"] == "m.text" && event.Sender != bot.client.UserID {
		if strings.HasPrefix(event.Content["body"].(string), "!problems") {
			problems := strings.Join(bot.getProblems(), "\n")
			bot.client.SendMessage(event.RoomID, problems)
		}
	}
}

func (bot ZabbixBot) initialSync() {
	resp := bot.client.InitialSync()
	for roomID := range resp.Rooms.Invite {
		bot.client.JoinRoom(roomID)
		log.Print("Joined room " + roomID)
	}
}

// Run runs the initial sync from the Matrix homeserver and begins processing events.
//
// This method does not return unless a fatal error occurs
func (bot ZabbixBot) Run() error {
	bot.initHTTP()
	bot.initialSync()
	return bot.client.Sync()
}

// NewZabbixBot creates a new ZabbixBot instance and initializes a matrix client
func NewZabbixBot(homeserverURL, userID, accessToken, zabbixAPIURL, zabbixUsername, zabbixPassword string, admin string) ZabbixBot {
	c := matrix.NewClient(homeserverURL, userID, accessToken)
	bot := ZabbixBot{
		c,
		make(map[string]string),
		zabbixAPIURL,
		zabbixUsername,
		zabbixPassword,
		admin,
	}
	c.OnEvent("m.room.member", bot.handleMemberEvent)
	c.OnEvent("m.room.message", bot.handleTextEvent)
	return bot
}
