package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
)

// matterMost struct holds data parsed via flags for the service
type matterMost struct {
	Title     string
	Token     string
	ServerURL string
	Scheme    string
	ApiURL    string
	Message   string
	ChanIDs   string
}

// SendToMattermost parse values from *cli.context and return *cli.Command
// and send messages to target channels.
// If multiple channel ids are provided then the string is split with "," separator and
// message is sent to each channel.
func SendToMattermost() *cli.Command {
	var mattermostOpts matterMost
	return &cli.Command{
		Name:  "mattermost",
		Usage: "Send message to mattermost",
		UsageText: "pingme mattermost --token '123' --channel '12345,567' --url 'localhost' --scheme http " +
			"--message 'some message'",
		Description: `Mattermost uses token to authenticate and channel ids for targets.
You can specify multiple channels by separating the value with ','.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &mattermostOpts.Token,
				Name:        "token",
				Aliases:     []string{"t"},
				Required:    true,
				Usage:       "Personal access token or Bot token for authorization.",
				EnvVars:     []string{"MATTERMOST_TOKEN"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.ChanIDs,
				Name:        "channel",
				Required:    false,
				Aliases:     []string{"c"},
				Usage:       "Channel IDs, if sending to multiple channels separate with ','.",
				EnvVars:     []string{"MATTERMOST_CHANNELS"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.Message,
				Name:        "msg",
				Aliases:     []string{"m"},
				Usage:       "Message content.",
				EnvVars:     []string{"MATTERMOST_MESSAGE"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.Title,
				Name:        "title",
				Value:       TimeValue,
				Usage:       "Title of the message.",
				EnvVars:     []string{"MATTERMOST_TITLE"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.ServerURL,
				Name:        "url",
				Aliases:     []string{"u"},
				Value:       "localhost",
				Required:    true,
				Usage:       "URL of your mattermost server i.e example.com",
				EnvVars:     []string{"MATTERMOST_SERVER_URL"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.Scheme,
				Name:        "scheme",
				Value:       "https",
				Usage:       "For server with no tls chose http, by default it uses https",
				EnvVars:     []string{"MATTERMOST_SCHEME"},
			},
			&cli.StringFlag{
				Destination: &mattermostOpts.ApiURL,
				Name:        "api",
				Value:       "/api/v4/posts",
				Usage:       "Unless using older version of api default is fine.",
				EnvVars:     []string{"MATTERMOST_API_URL"},
			},
		},
		Action: func(ctx *cli.Context) error {
			endPointURL := mattermostOpts.Scheme + "://" + mattermostOpts.ServerURL + mattermostOpts.ApiURL

			// Create a Bearer string by appending string access token
			bearer := "Bearer " + mattermostOpts.Token

			fullMessage := mattermostOpts.Title + "\n" + mattermostOpts.Message

			ids := strings.Split(mattermostOpts.ChanIDs, ",")
			for _, v := range ids {
				if len(v) == 0 {
					return fmt.Errorf(EmptyChannel)
				}

				jsonData, err := toJson(v, fullMessage)
				if err != nil {
					return fmt.Errorf("error parsing json\n[ERROR] - %v", err)
				}

				if err := sendMattermost(endPointURL, bearer, jsonData); err != nil {
					return fmt.Errorf("failed to send message\n[ERROR] - %v", err)
				}

			}
			return nil
		},
	}
}

// toJson takes strings and convert them to json byte array
func toJson(channel string, msg string) ([]byte, error) {
	if len(msg) == 0 {
		return nil, fmt.Errorf("Empty message")
	}
	m := make(map[string]string, 2)
	m["channel_id"] = channel
	m["message"] = msg
	js, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return js, nil
}

// sendMattermost function take the server url , authentication token
// message and channel id in the for of json byte array and sends
// message to mattermost vi http client.
func sendMattermost(url string, token string, jsonPayload []byte) error {
	// Create a new request using http
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	// add authorization header to the request
	r.Header.Set("Authorization", token)
	r.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Send request using http Client
	c := &http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	fmt.Println(string(body))
	resp.Body.Close()
	return nil
}