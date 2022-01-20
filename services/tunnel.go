package services

import (
	"errors"
	"net"
	"net/http"

	"github.com/koding/tunnel"
	"github.com/koding/tunnel/proto"
	"github.com/rs/zerolog/log"
)

// tunnel server

type globalTunnelHandler struct{}

func (gth globalTunnelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if tunnelServer != nil {
		tunnelServer.ServeHTTP(w, r)
	}
}

var TunnelHandler = globalTunnelHandler{}

var tunnelServer *tunnel.Server

func StartTunnelService() error {
	if tunnelServer != nil {
		return errors.New("there is already a tunnel server running")
	}

	var err error
	tunnelServer, err = tunnel.NewServer(&tunnel.ServerConfig{Debug: true})
	if err != nil {
		return err
	}

	return nil
}

func AddTunnelClient(subdomain string, identifier string) error {
	if tunnelServer == nil {
		return errors.New("no tunnel server")
	}

	tunnelServer.AddHost(subdomain+"."+TunnelDomain, identifier)
	return nil
}

func RemoveTunnelClient(subdomain string) error {
	if tunnelServer == nil {
		return errors.New("no tunnel server")
	}

	tunnelServer.DeleteHost(subdomain + "." + TunnelDomain)
	return nil
}

// tunnel client

type tunnellistener struct {
	remote net.Conn
}

func (tl tunnellistener) Accept() (net.Conn, error) { return tl.remote, nil }
func (tl tunnellistener) Close() error              { return nil }
func (tl tunnellistener) Addr() net.Addr            { return &tunneladdr{} }

type tunneladdr struct{}

func (ta tunneladdr) Network() string { return "tunnel" }
func (ta tunneladdr) String() string  { return "tunnel" }

func OpenTunnel(host string, identifier string) error {
	client, err := tunnel.NewClient(&tunnel.ClientConfig{
		Debug:      true,
		Identifier: identifier,
		ServerAddr: host,
		Proxy: func(remote net.Conn, msg *proto.ControlMessage) {
			if err := MainServer.Serve(tunnellistener{remote}); err != nil {
				log.Error().Err(err).Msg("error serving tunnel http")
			}
		},
	})
	if err != nil {
		return err
	}

	go client.Start()
	ok := client.StartNotify()
	<-ok
	return nil
}
