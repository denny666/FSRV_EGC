package googlehome

import (
	"context"
	"fmt"
	"net"
	"time"

	"FSRV_Edge/go-cast"
	"FSRV_Edge/go-cast/controllers"

	"github.com/evalphobia/google-tts-go/googletts"
)

// Client is Google Home client.
type Client struct {
	ctx    context.Context
	lang   string
	accent string
	ip     net.IP
	port   int
	Client *cast.Client
}

// NewClient creates Client from environment values.
func NewClient() (*Client, error) {
	return NewClientWithConfig(Config{})
}

// NewClientWithConfig creates Client from given Config.
func NewClientWithConfig(conf Config) (*Client, error) {
	host, err := conf.GetIPv4()
	if err != nil {
		return nil, err
	}
	port := conf.GetPort()
	client := cast.NewClient(host, port)

	ctx := conf.GetOrCreateContext()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{
		ctx:    ctx,
		ip:     host,
		port:   port,
		lang:   conf.GetLang(),
		accent: conf.GetAccent(),
		Client: client,
	}, nil
}

// SetLang sets lang.
func (c *Client) SetLang(lang string) {
	c.lang = lang
}

// SetAccent sets accent.
func (c *Client) SetAccent(accent string) {
	c.accent = accent
}

// GetIPv4 returns IPv4 address of Google Home.
func (c *Client) GetIPv4() string {
	return c.ip.String()
}

// GetContext GetContext
func (c *Client) GetContext() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Notify make Google Home say something interesting.
func (c *Client) Notify(text string, language ...string) error {
	lang := c.lang
	if len(language) != 0 {
		lang = language[0]
	}
	if c.accent != "" {
		lang = fmt.Sprintf("%s-%s", lang, c.accent)
	}

	url, err := googletts.GetTTSURL(text, lang)
	if err != nil {
		return err
	}
	return c.Play(url)
}

// Play make Google Home play music or sound.
func (c *Client) Play(url string) error {
	client := cast.NewClient(c.ip, c.port)
	defer client.Close()
	err := client.Connect(c.ctx)
	if err != nil {
		return err
	}
	client.Receiver().QuitApp(c.ctx)

	media, err := client.Media(c.ctx)
	if err != nil {
		return err
	}

	item := controllers.MediaItem{
		ContentID:  url,
		StreamType: "LIVE",
		//audio/mpeg
		ContentType: "audio/mpeg",
	}
	_, err = media.LoadMedia(c.ctx, item, 0, true, map[string]interface{}{})
	return err
}

// URL URL
func (c *Client) URL(url, playType string) error {
	client := cast.NewClient(c.ip, c.port)
	defer client.Close()
	err := client.Connect(c.ctx)
	if err != nil {
		return err
	}
	client.Receiver().QuitApp(c.ctx)

	media, err := client.URL(c.ctx)
	if err != nil {
		return err
	}

	_, err = media.LoadURL(c.ctx, url, playType)
	return err
}

// QuitApp stops recveiver application.
func (c *Client) QuitApp() error {
	client := cast.NewClient(c.ip, c.port)
	defer client.Close()

	connectErr := client.Connect(c.ctx)
	if connectErr != nil {
		return connectErr
	}
	receiver := client.Receiver()
	_, err := receiver.QuitApp(c.ctx)
	return err
}

// GetStatus gets volume.
func (c *Client) GetStatus() (receiverStatus *controllers.ReceiverStatus, err error) {
	client := cast.NewClient(c.ip, c.port)
	defer client.Close()
	err = client.Connect(c.ctx)
	if err != nil {
		return receiverStatus, err
	}
	receiver := client.Receiver()

	if receiver == nil {
		var errMsg ErrMsg
		errMsg.SetMsg("no receiver")
		return receiverStatus, errMsg
	}

	status, err := receiver.GetStatus(c.ctx)
	if err != nil {
		return receiverStatus, err
	}

	return status, nil
}

// ErrMsg error message
type ErrMsg struct {
	msg string
}

//SetMsg 設置msg
func (err ErrMsg) SetMsg(msg string) {
	err.msg = msg
}

func (err ErrMsg) Error() string {
	return err.msg
}
