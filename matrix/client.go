package matrix

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/matrix-org/gomatrix"
)

type Client struct {
	UserID         string
	client         *gomatrix.Client
	outboundEvents chan outboundEvent
}

type outboundEvent struct {
	RoomID         string
	EventType      string
	Content        interface{}
	RetryOnFailure bool
	done           chan<- string
}

type simpleMessage struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
}

type messageEdit struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	Format        string `json:"format"`
	FormattedBody string `json:"formatted_body"`
	NewContent    struct {
		MsgType       string `json:"msgtype"`
		Body          string `json:"body"`
		Format        string `json:"format"`
		FormattedBody string `json:"formatted_body"`
	} `json:"m.new_content"`
	RelatesTo struct {
		RelType string `json:"rel_type"`
		EventID string `json:"event_id"`
	} `json:"m.relates_to"`
}

type httpError struct {
	Errcode      string `json:"errcode"`
	Err          string `json:"error"`
	RetryAfterMs int    `json:"retry_after_ms"`
}

func (c Client) sendMessage(roomID string, message interface{}, retryOnFailure bool) <-chan string {
	done := make(chan string, 1)
	c.outboundEvents <- outboundEvent{roomID, "m.room.message", message, retryOnFailure, done}
	return done
}

// InitialSync gets the initial sync from the server for catching up with important missed event such as invites
func (c Client) InitialSync() *gomatrix.RespSync {
	resp, err := c.client.SyncRequest(0, "", "", false, "")
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

// Sync begins synchronizing the events from the server and returns only in case of a severe error
func (c Client) Sync() error {
	return c.client.Sync()
}

func (c Client) OnEvent(eventType string, callback gomatrix.OnEventListener) {
	c.client.Syncer.(*gomatrix.DefaultSyncer).OnEventType(eventType, callback)
}

func (c Client) JoinRoom(roomID string) {
	_, err := c.client.JoinRoom(roomID, "", nil)
	if err != nil {
		log.Print("Failed to join room "+roomID+": ", err)
	}
}

// SendMessage queues a message to be sent and returns immediatedly.
//
// The returned channel will provide the event ID of the message after the message has been sent
func (c Client) SendMessage(roomID string, message string) <-chan string {
	return c.sendMessage(roomID, simpleMessage{"m.notice", strip.StripTags(message), "org.matrix.custom.html", strings.ReplaceAll(message, "\n", "<br />")}, true)
}

// EditMessage edits a previously sent message identified by its event ID
func (c Client) EditMessage(roomID string, eventID string, message string) <-chan string {
	msgEdit := messageEdit{}
	msgEdit.Body = strip.StripTags(message)
	msgEdit.FormattedBody = strings.ReplaceAll(message, "\n", "<br />")
	msgEdit.Format = "org.matrix.custom.html"
	msgEdit.NewContent.Body = strip.StripTags(message)
	msgEdit.NewContent.FormattedBody = strings.ReplaceAll(message, "\n", "<br />")
	msgEdit.NewContent.Format = "org.matrix.custom.html"
	msgEdit.MsgType = "m.notice"
	msgEdit.NewContent.MsgType = "m.notice"
	msgEdit.RelatesTo.RelType = "m.replace"
	msgEdit.RelatesTo.EventID = eventID
	return c.sendMessage(roomID, msgEdit, true)
}

// SendStreamingMessage creates a pair of channels that can be used to send and update (by editing) a message in place.
//
// The initial message will be sent when messageUpdate receives the first message. The message will be
// updated until done is closed, at which point messageUpdate will be drained and the last version be updated.
func (c Client) SendStreamingMessage(roomID string) (messageUpdate chan<- string, done chan<- struct{}) {
	input := make(chan string, 256)
	doneChan := make(chan struct{})
	go func() {
		text := <-input
		id := <-c.SendMessage(roomID, text)
		msgEdit := messageEdit{}
		msgEdit.Body = strip.StripTags(text)
		msgEdit.FormattedBody = strings.ReplaceAll(text, "\n", "<br />")
		msgEdit.Format = "org.matrix.custom.html"
		msgEdit.NewContent.Body = strip.StripTags(text)
		msgEdit.NewContent.FormattedBody = strings.ReplaceAll(text, "\n", "<br />")
		msgEdit.NewContent.Format = "org.matrix.custom.html"
		msgEdit.MsgType = "m.notice"
		msgEdit.NewContent.MsgType = "m.notice"
		msgEdit.RelatesTo.RelType = "m.replace"
		msgEdit.RelatesTo.EventID = id
		messageDone := false
		for !messageDone {
			select { // Wait for more input or done signal
			case m := <-input:
				msgEdit.Body = strip.StripTags(m)
				msgEdit.FormattedBody = strings.ReplaceAll(m, "\n", "<br />")
				msgEdit.NewContent.Body = strip.StripTags(m)
				msgEdit.NewContent.FormattedBody = strings.ReplaceAll(m, "\n", "<br />")
			case <-doneChan:
				messageDone = true
			}
			for messages := true; messages; { // drain the input in case done was signaled
				select {
				case m := <-input:
					msgEdit.Body = strip.StripTags(m)
					msgEdit.FormattedBody = strings.ReplaceAll(m, "\n", "<br />")
					msgEdit.NewContent.Body = strip.StripTags(m)
					msgEdit.NewContent.FormattedBody = strings.ReplaceAll(m, "\n", "<br />")
				default:
					messages = false
				}
			}
			res := <-c.sendMessage(roomID, msgEdit, messageDone)
			if res == "" { // no event id, send failed, wait for a bit before retrying
				time.Sleep(100 * time.Millisecond)
			}
		}

	}()
	return input, doneChan
}

func processOutboundEvents(client Client) {
	for event := range client.outboundEvents {
		for {
			resp, err := client.client.SendMessageEvent(event.RoomID, event.EventType, event.Content)
			if err == nil {
				if event.done != nil {
					event.done <- resp.EventID
				}
				break // Success, break the retry loop
			}
			var httpErr httpError
			if jsonErr := json.Unmarshal(err.(gomatrix.HTTPError).Contents, &httpErr); jsonErr != nil {
				log.Print("Failed to parse error response!", jsonErr)
			}

			fatalFailure := false

			switch e := httpErr.Errcode; e {
			case "M_LIMIT_EXCEEDED":
				time.Sleep(time.Duration(httpErr.RetryAfterMs) * time.Millisecond)
			case "M_FORBIDDEN":
				event.done <- ""
				fatalFailure = true
				fallthrough
			default:
				log.Print("Failed to send message to room "+event.RoomID+" err: ", err)
				log.Print(string(err.(gomatrix.HTTPError).Contents))
			}
			if !event.RetryOnFailure || fatalFailure {
				event.done <- ""
				break
			}
		}
	}
}

// NewClient creates a new Matrix client and performs basic initialization on it
func NewClient(homeserverURL, userID, accessToken string) Client {
	client, err := gomatrix.NewClient(homeserverURL, userID, accessToken)
	if err != nil {
		log.Fatal(err)
	}
	c := Client{
		userID,
		client,
		make(chan outboundEvent, 256),
	}
	go processOutboundEvents(c)
	return c
}
