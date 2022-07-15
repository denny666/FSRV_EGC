package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"FSRV_Edge/go-cast/api"
	"FSRV_Edge/go-cast/events"
	"FSRV_Edge/go-cast/log"
	"FSRV_Edge/go-cast/net"
)

type YoutubeController struct {
	interval      time.Duration
	channel       *net.Channel
	eventsCh      chan events.Event
	DestinationID string
	URLSessionID  int
}

const NamespaceYoutube = "urn:x-cast:com.google.youtube.mdx"

var getYoutubeStatus = net.PayloadHeaders{Type: "GET_STATUS"}

var commandYoutubeLoad = net.PayloadHeaders{Type: "LOAD"}

type LoadYoutubeCommand struct {
	net.PayloadHeaders
	Url       string `json:url`
	Type      string `json:"type"`
	Data      d      `json:"data"`
	RequestId int    `json:"requestId"`
}
type d struct {
	CurrentTime int64  `json:"currentTime`
	VideoId     string `json:"videoId`
}
type YoutubeStatusURL struct {
	ContentId   string  `json:"contentId"`
	StreamType  string  `json:"streamType"`
	ContentType string  `json:"contentType"`
	Duration    float64 `json:"duration"`
}

func NewYoutubeController(conn *net.Connection, eventsCh chan events.Event, sourceId, destinationID string) *YoutubeController {
	controller := &YoutubeController{
		channel:       conn.NewChannel(sourceId, destinationID, NamespaceURL),
		eventsCh:      eventsCh,
		DestinationID: destinationID,
	}

	controller.channel.OnMessage("URL_STATUS", controller.onStatus)

	return controller
}

func (c *YoutubeController) SetDestinationID(id string) {
	c.channel.DestinationId = id
	c.DestinationID = id
}

func (c *YoutubeController) sendEvent(event events.Event) {
	select {
	case c.eventsCh <- event:
	default:
		log.Printf("Dropped event: %#v", event)
	}
}

func (c *YoutubeController) onStatus(message *api.CastMessage) {
	response, err := c.parseStatus(message)
	if err != nil {
		log.Errorf("Error parsing status: %s", err)
	}

	for _, status := range response.Status {
		c.sendEvent(*status)
	}
}

func (c *YoutubeController) parseStatus(message *api.CastMessage) (*URLStatusResponse, error) {
	response := &URLStatusResponse{}

	err := json.Unmarshal([]byte(*message.PayloadUtf8), response)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal status message:%s - %s", err, *message.PayloadUtf8)
	}

	for _, status := range response.Status {
		c.URLSessionID = status.URLSessionID
	}

	return response, nil
}

type YoutubeStatusResponse struct {
	net.PayloadHeaders
	Status []*URLStatus `json:"status,omitempty"`
}

type YoutubeStatus struct {
	net.PayloadHeaders
	URLSessionID         int                    `json:"mediaSessionId"`
	PlaybackRate         float64                `json:"playbackRate"`
	PlayerState          string                 `json:"playerState"`
	CurrentTime          float64                `json:"currentTime"`
	SupportedURLCommands int                    `json:"supportedURLCommands"`
	Volume               *Volume                `json:"volume,omitempty"`
	URL                  *URLStatusURL          `json:"media"`
	CustomData           map[string]interface{} `json:"customData"`
	RepeatMode           string                 `json:"repeatMode"`
	IdleReason           string                 `json:"idleReason"`
}

func (c *YoutubeController) Start(ctx context.Context) error {
	_, err := c.GetStatus(ctx)
	return err
}

func (c *YoutubeController) GetStatus(ctx context.Context) (*URLStatusResponse, error) {
	message, err := c.channel.Request(ctx, &getURLStatus)
	if err != nil {
		return nil, fmt.Errorf("Failed to get receiver status: %s", err)
	}

	return c.parseStatus(message)
}

func (c *YoutubeController) LoadYoutube(ctx context.Context, url string) (*api.CastMessage, error) {
	var data d
	data.CurrentTime = 0
	data.VideoId = "JMl8cQjBfqk"
	message, err := c.channel.Request(ctx, &LoadYoutubeCommand{
		PayloadHeaders: commandURLLoad,
		Type:           "flingVideo",
		Data:           data,
		RequestId:      1,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to send load command: %s", err)
	}

	response := &net.PayloadHeaders{}
	err = json.Unmarshal([]byte(*message.PayloadUtf8), response)
	if err != nil {
		return nil, err
	}
	if response.Type == "LOAD_FAILED" {
		return nil, errors.New("Load URL failed")
	}

	return message, nil
}
