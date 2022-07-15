package controllers

import (
	"FSRV_Edge/go-cast/events"

	"FSRV_Edge/go-cast/net"

	"golang.org/x/net/context"
)

type ConnectionController struct {
	channel *net.Channel
}

var connect = net.PayloadHeaders{Type: "CONNECT"}
var close = net.PayloadHeaders{Type: "CLOSE"}

func NewConnectionController(conn *net.Connection, eventsCh chan events.Event, sourceId, destinationId string) *ConnectionController {
	controller := &ConnectionController{
		channel: conn.NewChannel(sourceId, destinationId, "urn:x-cast:com.google.cast.tp.connection"),
	}

	return controller
}

func (c *ConnectionController) Start(ctx context.Context) error {
	return c.channel.Send(connect)
}

func (c *ConnectionController) Close() error {
	return c.channel.Send(close)
}
