package bot

import (
	"strconv"

	"github.com/cavaliercoder/go-zabbix"
)

func (bot ZabbixBot) getProblems() []string {
	session, err := zabbix.NewSession(bot.zabbixAPIURL, bot.zabbixUsername, bot.zabbixPassword)
	if err != nil {
		return []string{err.Error()}
	}

	params := zabbix.TriggerGetParams{}
	params.RecentProblemOnly = true
	params.ActiveOnly = true
	params.ExpandDescription = true
	params.SelectHosts = []string{"host"}
	triggers, err := session.GetTriggers(params)
	if err != nil {
		return []string{err.Error()}
	}

	problems := []string{strconv.Itoa(len(triggers)) + " active problems:"}
	for _, trigger := range triggers {
		problems = append(problems, trigger.Description+" (Host: <code>"+trigger.Hosts[0].Hostname+"</code>)")
	}
	return problems
}
