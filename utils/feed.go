package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mmcdole/gofeed"
)

var fp = gofeed.NewParser()

func ParseFeed(feedURL string) (map[string]interface{}, string, error) {
	resp, err := http.Get(feedURL)
	if err != nil {
		return nil, "", err
	}
	if resp != nil {
		defer func() {
			ce := resp.Body.Close()
			if ce != nil {
				err = ce
			}
		}()
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("got bad status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	bodyString := string(body)
	feed, err := fp.ParseString(bodyString)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing feed: %w", err)
	}

	var mapfeed map[string]interface{}
	j, _ := json.Marshal(feed)
	json.Unmarshal(j, &mapfeed)

	return mapfeed, bodyString, err
}
