package bot

// AlertMessage as received from a Zabbix alertscript
type AlertMessage struct {
	SendTo  string `json:"send_to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// AlertUpdate updates an alert by editing an existing message with the same subject (=ID) or posting a new
func (bot ZabbixBot) AlertUpdate(msg AlertMessage) {
	eventID := bot.zabbixIdentifierToMatrixEvent[msg.SendTo+":"+msg.Subject]
	if eventID == "" {
		eventID := <-bot.client.SendMessage(msg.SendTo, msg.Message)
		bot.zabbixIdentifierToMatrixEvent[msg.SendTo+":"+msg.Subject] = eventID
	} else {
		bot.client.EditMessage(msg.SendTo, eventID, msg.Message)
	}
}
