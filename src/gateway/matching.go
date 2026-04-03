package gateway

import "net/http"

type MatchingGateway struct {
	URL    string
	Client *http.Client
}

func NewMatchingGateway(url string) *MatchingGateway {
	return &MatchingGateway{
		URL:    url,
		Client: http.DefaultClient,
	}
}

func (g *MatchingGateway) Match() {}
