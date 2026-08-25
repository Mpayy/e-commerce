package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Mpayy/e-commerce/services/api-gateway/internal/gateway/config"
)

type Gateway struct {
	routes  []config.Route
	proxies map[string]*httputil.ReverseProxy
}

func NewGateway(routes []config.Route) (*Gateway, error) {
	proxies := make(map[string]*httputil.ReverseProxy)

	for _, route := range routes {
		if _, exists := proxies[route.Target]; exists {
			continue
		}

		target, err := url.Parse(route.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid target %q: %w", route.Target, err)
		}

		proxies[route.Target] = httputil.NewSingleHostReverseProxy(target)
	}

	return &Gateway{
		routes:  routes,
		proxies: proxies,
	}, nil
}

func (g *Gateway) Handler() http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range g.routes {
			if strings.HasPrefix(r.URL.Path, route.PathPrefix) {
				g.proxies[route.Target].ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"error":"route not found"}`))
	})

	return handler
}
