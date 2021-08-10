package kubernetes

import (
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
)

type edgeTransport struct {
	*baseTransport
	reverseTunnelService portainer.ReverseTunnelService
}

// NewAgentTransport returns a new transport that can be used to send signed requests to a Portainer Edge agent
func NewEdgeTransport(dataStore portainer.DataStore, reverseTunnelService portainer.ReverseTunnelService, endpoint *portainer.Endpoint, tokenManager *tokenManager) *edgeTransport {
	transport := &edgeTransport{
		reverseTunnelService: reverseTunnelService,
		baseTransport: newBaseTransport(
			&http.Transport{},
			tokenManager,
			endpoint,
			dataStore,
		),
	}

	return transport
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *edgeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	token, err := getRoundTripToken(request, transport.tokenManager, transport.endpoint.ID)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portainer.PortainerAgentKubernetesSATokenHeader, token)

	if strings.HasPrefix(request.URL.Path, "/v2") {
		decorateAgentRequest(request, transport.dataStore)
	}

	response, err := transport.baseTransport.RoundTrip(request)

	if err == nil {
		transport.reverseTunnelService.SetTunnelStatusToActive(transport.endpoint.ID)
	} else {
		transport.reverseTunnelService.SetTunnelStatusToIdle(transport.endpoint.ID)
	}

	return response, err
}
