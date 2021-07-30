package utils

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"msmf/database"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Message    string
	RetryAfter float64
	global     bool
}

// newRequest makes the request object and reader
func newRequest(url string, body *map[string]string) (req *http.Request, err error) {
	// Create a reader for the body
	reader := bytes.NewReader(ToJSON(body))

	// Create the request
	req, err = http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}

	// Set appropriate headers
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func sendRequest(client *http.Client, url string, count *int, body *map[string]string) (waitUntil time.Time, err error) {
	// Loop waitUntil we get a good status code
	for {
		req, err := newRequest(url, body)

		// See if we need to wait again before sending the request
		if !waitUntil.IsZero() {
			// Sleep until that time
			time.Sleep(time.Until(waitUntil))
		}

		// Send the request
		resp, err := client.Do(req)
		now := time.Now()
		if err != nil {
			return now.Add(500 * time.Millisecond), err
		}

		// This is good
		if resp.StatusCode == http.StatusNoContent {
			// If there is no count on waiting, just set it to standard 500ms wait time
			if *count == 0 {
				return now.Add(500 * time.Millisecond), nil
			} else if *count > 0 {
				// Drop the count by 1
				*count -= 1
			}
			return now.Add(500 * time.Millisecond), nil
		}

		// If sending things way too fast
		if resp.StatusCode == http.StatusTooManyRequests {
			// Get headers
			buffer := bytes.Buffer{}
			_ = resp.Header.Write(&buffer)

			// Get the amount of requests remaining
			remainingStr := resp.Header.Get("X-Ratelimit-Remaining")
			remaining, err := strconv.Atoi(remainingStr)
			if err != nil {
				return now.Add(500 * time.Millisecond), err
			}

			// Get the unix timestamp when the bucket will reset
			resetStr := resp.Header.Get("X-Ratelimit-Reset")
			resetInt, err := strconv.Atoi(resetStr)
			if err != nil {
				return now.Add(500 * time.Millisecond), err
			}

			// Calculate time between now and when the token bucket resets
			reset := time.Unix(int64(resetInt), 0)
			now := time.Now()
			waitUntil = now.Add(now.Sub(reset) / time.Duration(remaining+1))

			// Set count to remaining
			*count = remaining

			// Get JSON of body
			data := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				return waitUntil, err
			}

			time.Sleep(time.Duration(data["retry_after"].(float64)) * time.Millisecond)
		}
	}
}

func RunDiscordIntegration(connDetails *ConnDetails, serverID int) {
	// See if there is a discord integration
	var integration database.DiscordIntegration
	err := database.DB.Where("server_id = ?", serverID).Find(&integration).Error
	// If there is no error continue
	if err == nil {
		// See if we are supposed to use it and that it's a webhook
		if integration.Active && integration.Type == "webhook" {
			// Create a fake connection
			discord := &websocket.Conn{}
			discord = nil

			// Make the channels for stdin and stdout
			pipes := PipeChans{
				// A lot of messages might clog this up because of rate limiting
				StdoutChan: make(chan []byte, 20),
				StderrChan: make(chan []byte, 5),
			}

			// Register with the SPMC and NoRepeat
			connDetails.SLock.Lock()
			connDetails.SPMC[discord] = pipes
			connDetails.NoRepeat[discord] = pipes
			connDetails.SLock.Unlock()

			// Go handle the integration
			go SendWebhook(integration, pipes)
		}
	}
}

// StopDiscordIntegration will kill a running integration
func StopDiscordIntegration(serverID int) {
	connDetails := AttachServer(serverID)

	// Get the pipes to the Discord integration connection and remove it from the map
	connDetails.SLock.Lock()
	pipes, _ := connDetails.NoRepeat[nil]
	delete(connDetails.NoRepeat, nil)
	pipes, exists := connDetails.SPMC[nil]
	delete(connDetails.SPMC, nil)
	connDetails.SLock.Unlock()

	// Close the pipes so they stop trying to read
	if exists {
		close(pipes.StdoutChan)
		close(pipes.StderrChan)
	}
}

// SendWebhook will send over the data to Discord
func SendWebhook(integration database.DiscordIntegration, pipes PipeChans) {
	// Create a new http client
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Create a new rate limiter
	count := 0
	// Standard wait time of 500ms
	waitUntil := time.Now()
	var err error

	// Forever look for messages
	for {
		body := make(map[string]string)
		// Add avatar if we have it
		if integration.AvatarURL != nil {
			body["avatar_url"] = *integration.AvatarURL
		}
		// Add username if we have it
		if integration.Username != nil {
			body["username"] = *integration.Username
		}

		var data []byte
		ok := true

		// Make a timer to batch until this goes off
		timer := time.NewTimer(time.Until(waitUntil))

		// Get data from the channels
		canSend := false
		// TODO make this more robust to only handle either stdout or stderr at a time
		// once one is added to the queue. Also make sure content doesn't go past
		// the 2000 character count
		for {
			select {
			case data, ok = <-pipes.StdoutChan:
				body["content"] += string(data)

				// Timer has already gone off so we can send this on
				if canSend {
					canSend = false
					goto send
				}
			case data, ok = <-pipes.StderrChan:
				if len(body["content"]) == 0 {
					body["content"] += "ERROR: " + string(data)
				} else {
					body["content"] += string(data)
				}

				// Timer has already gone off so we can send this on
				if canSend {
					canSend = false
					goto send
				}
			// Time is up and we can send the message
			case <-timer.C:
				// Time to break and send whatever is in the body already
				if len(body["content"]) > 0 {
					goto send
				} else {
					// Nothing is in the body already, so wait for it
					canSend = true
				}
			}

			// Our pipes are closed, so end function
			if !ok {
				return
			}
		}

	send:
		// Send the request
		waitUntil, err = sendRequest(client, integration.DiscordURL, &count, &body)
		if err != nil {
			log.Println("Webhook error:", err)
			return
		}
	}
}
