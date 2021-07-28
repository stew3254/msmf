package utils

import (
	"bytes"
	"encoding/json"
	"github.com/stew3254/ratelimit"
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

func sendRequest(
	client *http.Client,
	url string,
	rl *ratelimit.RateLimiter,
	count *int,
	body *map[string]string,
) error {
	// Loop until we get a good status code
	for {
		req, err := newRequest(url, body)

		// Send the request
		rl.Lock()
		resp, err := client.Do(req)
		rl.Unlock()
		if err != nil {
			return err
		}

		// This is good
		if resp.StatusCode == http.StatusNoContent {
			// If there is no count on waiting, just set it to standard 500ms wait time
			if *count == 0 {
				// Set the wait limit to 500ms
				rl.SetLimit(500 * time.Millisecond)
			} else if *count > 0 {
				// Drop the count by 1
				*count -= 1
			}
			return nil
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
				return err
			}

			// Get the unix timestamp when the bucket will reset
			resetStr := resp.Header.Get("X-Ratelimit-Reset")
			resetInt, err := strconv.Atoi(resetStr)
			if err != nil {
				return err
			}

			// Calculate time between now and when the token bucket resets
			reset := time.Unix(int64(resetInt), 0)
			now := time.Now()
			wait := now.Sub(reset) / time.Duration(remaining+1)

			// Set the wait limit for the next requests
			rl.SetLimit(wait)
			// Set count to remaining
			*count = remaining

			// Get JSON of body
			data := make(map[string]interface{})
			err = json.NewDecoder(resp.Body).Decode(&data)
			if err != nil {
				return err
			}

			log.Println("Wait for", data["retry_after"])
			time.Sleep(time.Duration(data["retry_after"].(float64)) * time.Millisecond)
		}
	}
}

// SendWebhook will send over the data to Discord
func SendWebhook(integration database.DiscordIntegration, pipes PipeChans) {
	// Create a new http client
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Create a new rate limiter
	rl := ratelimit.NewRateLimiter(1, 5, 500*time.Millisecond, time.Millisecond)
	count := 0

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
		var ok bool

		// Get data from the channels
		select {
		case data, ok = <-pipes.StdoutChan:
			body["content"] = string(data)
		case data, ok = <-pipes.StderrChan:
			body["content"] = "ERROR: " + string(data)
		}
		// Our pipes are closed, so break out
		if !ok {
			return
		}

		// Send the request
		err := sendRequest(client, integration.DiscordURL, rl, &count, &body)
		if err != nil {
			log.Println("Webhook error:", err)
			return
		}
	}
}
