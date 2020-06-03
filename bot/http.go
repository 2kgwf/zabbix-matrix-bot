package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (bot ZabbixBot) initHTTP() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		msg := AlertMessage{}
		err := json.NewDecoder(req.Body).Decode(&msg)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
		req.Body.Close()
		bot.AlertUpdate(msg)
		fmt.Fprintf(w, "OK")
	})
	go http.ListenAndServe(":8080", nil)
}
