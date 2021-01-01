package server

import (
	"net/http"

	"not-for-work/aviasales_test/cache"
	"not-for-work/aviasales_test/store"
)

type Server struct {
	Cache *cache.Cache
	Store *store.Store
}

func New(c *cache.Cache, s *store.Store) *http.ServeMux {
	h := &Server{
		Cache: c,
		Store: s,
	}

	router := http.NewServeMux()

	router.HandleFunc("/get", h.GetHandler)
	router.HandleFunc("/load", h.LoadHandler)

	return router
}
