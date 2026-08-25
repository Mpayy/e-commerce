package config

import (
	"sort"
)

type ServiceTargets struct {
	UserServiceAddr    string
	ProductServiceAddr string
	OrderServiceAddr   string
}

type Route struct {
	PathPrefix string
	Target     string
}

func BuildRoutes(targets ServiceTargets) []Route {
	routes := []Route{
		{PathPrefix: "/api/v1/admin/products", Target: targets.ProductServiceAddr},
		{PathPrefix: "/api/v1/admin/categories", Target: targets.ProductServiceAddr},
		{PathPrefix: "/api/v1/products", Target: targets.ProductServiceAddr},
		{PathPrefix: "/api/v1/categories", Target: targets.ProductServiceAddr},

		{PathPrefix: "/api/v1/register", Target: targets.UserServiceAddr},
		{PathPrefix: "/api/v1/login", Target: targets.UserServiceAddr},
		{PathPrefix: "/api/v1/profile", Target: targets.UserServiceAddr},
		{PathPrefix: "/api/v1/logout", Target: targets.UserServiceAddr},

		{PathPrefix: "/api/v1/cart", Target: targets.OrderServiceAddr},
		{PathPrefix: "/api/v1/orders", Target: targets.OrderServiceAddr},
	}

	sort.Slice(routes, func(i, j int) bool {
		result := len(routes[i].PathPrefix) > len(routes[j].PathPrefix)
		return result
	})

	return routes
}
