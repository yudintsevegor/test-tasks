package server

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
	"time"

	"not-for-work/aviasales_test/cache"
	"not-for-work/aviasales_test/config"
	"not-for-work/aviasales_test/internal/key_generation"
	"not-for-work/aviasales_test/store"

	"github.com/stretchr/testify/assert"
)

func Test_GetHandler(t *testing.T) {
	tt := assert.New(t)

	cfg, err := config.Load("./../config/test.env")
	tt.NoError(err)
	log.Printf("CONFIG: %+v", cfg)

	c, err := cache.New(cfg.Redis, cfg.DropData)
	tt.NoError(err)
	defer c.Close()

	s, err := store.New(cfg.Mongo, cfg.DropData)
	tt.NoError(err)
	defer s.Close(context.Background())

	ts := httptest.NewServer(New(c, s))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	tt.NoError(err)

	cases := []struct {
		caseName         string
		actionBeforeTest func() error
		reqQueryParams   map[string][]string
		expectedCode     int
		expectedBody     []byte
	}{
		{
			caseName:         "many query params",
			actionBeforeTest: func() error { return nil },
			reqQueryParams:   map[string][]string{wordQueryParam: {"foobar", "barfoo"}},
			expectedCode:     http.StatusBadRequest,
			expectedBody:     []byte(""),
		},
		{
			caseName:         "bad query param",
			actionBeforeTest: func() error { return nil },
			reqQueryParams:   map[string][]string{wordQueryParam + "test": {"foobar"}},
			expectedCode:     http.StatusBadRequest,
			expectedBody:     []byte(""),
		},
		{
			caseName:         "empty stores",
			actionBeforeTest: func() error { return nil },
			reqQueryParams:   map[string][]string{wordQueryParam: {"foobar"}},
			expectedCode:     http.StatusNotFound,
			expectedBody:     []byte(""),
		},
		{
			caseName: "get data from store",
			// we need add something to database before case
			actionBeforeTest: func() error {
				o := store.Object{
					Key:   key_generation.New("foobar"),
					Words: []string{"foobar"},
				}
				_, err := s.Put(context.Background(), o)
				return err
			},
			reqQueryParams: map[string][]string{wordQueryParam: {"foobar"}},
			expectedCode:   http.StatusOK,
			expectedBody:   []byte(`["foobar"]`),
		},
		{
			caseName:         "get data from cache",
			actionBeforeTest: func() error { return nil },
			reqQueryParams:   map[string][]string{wordQueryParam: {"foobar"}},
			expectedCode:     http.StatusOK,
			expectedBody:     []byte(`["foobar"]`),
		},
		{
			caseName: "get data from store (wait cache flush)",
			// we need add something to cache before case
			actionBeforeTest: func() error {
				// wait empty cache
				time.Sleep(time.Duration(cfg.Redis.CacheTTL+1) * time.Second)
				return c.Put(context.Background(), key_generation.New("foobar"), []string{"foobar"}...)
			},
			reqQueryParams: map[string][]string{wordQueryParam: {"foobar"}},
			expectedCode:   http.StatusOK,
			expectedBody:   []byte(`["foobar"]`),
		},
		{
			caseName: "add empty words to store",
			// we need add something to store before case
			actionBeforeTest: func() error {
				o := store.Object{
					Key:   key_generation.New("test"),
					Words: nil,
				}
				_, err := s.Put(context.Background(), o)
				return err
			},
			reqQueryParams: map[string][]string{wordQueryParam: {"test"}},
			expectedCode:   http.StatusNotFound,
			expectedBody:   []byte(""),
		},
	}

	for i, c := range cases {
		tt.NoError(c.actionBeforeTest())

		reqURL := createReqURL(*u, "get", c.reqQueryParams)
		resp, err := http.Get(reqURL)
		tt.NoError(err)

		message := "[%d] case name: %s, URL: %s"
		tt.Equalf(c.expectedCode, resp.StatusCode, message, i, c.caseName, reqURL)

		body, err := ioutil.ReadAll(resp.Body)
		tt.NoError(err)

		tt.Equalf(string(c.expectedBody), string(body), message, i, c.caseName, reqURL)
	}
}

func Test_LoadHandler(t *testing.T) {
	tt := assert.New(t)

	cfg, err := config.Load("./../config/test.env")
	tt.NoError(err)
	log.Printf("CONFIG: %+v", cfg)

	c, err := cache.New(cfg.Redis, cfg.DropData)
	tt.NoError(err)
	defer c.Close()

	s, err := store.New(cfg.Mongo, cfg.DropData)
	tt.NoError(err)
	defer s.Close(context.Background())

	ts := httptest.NewServer(New(c, s))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	tt.NoError(err)

	cases := []struct {
		caseName         string
		actionBeforeTest func() error
		reqBody          []byte
		expectedCode     int
		expectedBody     []byte
	}{
		{
			caseName:         "bad body",
			actionBeforeTest: func() error { return nil },
			reqBody:          []byte("{}"),
			expectedCode:     http.StatusBadRequest,
			expectedBody:     []byte(""),
		},
		{
			caseName:         "empty req body",
			actionBeforeTest: func() error { return nil },
			reqBody:          nil,
			expectedCode:     http.StatusBadRequest,
			expectedBody:     []byte(""),
		},
		{
			caseName: "not uniq words",
			actionBeforeTest: func() error {
				return c.Put(context.Background(), key_generation.New("foobar"), []string{"foobar"}...)
			},
			reqBody:      []byte(`["foobar"]`),
			expectedCode: http.StatusOK,
			expectedBody: []byte(""),
		},
	}

	for i, c := range cases {
		tt.NoError(c.actionBeforeTest())

		reqURL := createReqURL(*u, "load", nil)
		resp, err := http.Post(reqURL, "application/json", bytes.NewReader(c.reqBody))
		tt.NoError(err)

		message := "[%d] case name: %s, URL: %s"
		tt.Equalf(c.expectedCode, resp.StatusCode, message, i, c.caseName, reqURL)

		body, err := ioutil.ReadAll(resp.Body)
		tt.NoError(err)

		tt.Equalf(string(c.expectedBody), string(body), message, i, c.caseName, reqURL)
	}
}

func createReqURL(u url.URL, reqPath string, reqParams map[string][]string) string {
	u.Path = path.Join(u.Path, reqPath)

	query := make(url.Values)
	for k, values := range reqParams {
		for _, v := range values {
			query.Add(k, v)
		}
	}
	u.RawQuery = query.Encode()

	return u.String()
}
