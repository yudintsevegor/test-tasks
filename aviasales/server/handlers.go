package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"not-for-work/aviasales_test/cache"
	keygen "not-for-work/aviasales/internal/key_generation"
	"not-for-work/aviasales_test/store"
)

const wordQueryParam = "word"

func (s *Server) GetHandler(rw http.ResponseWriter, r *http.Request) {
	notify := func(err error, code int) {
		log.Printf("GetHandler: %v", err)
		rw.WriteHeader(code)
	}

	word, err := getQueryParam(r.URL.Query(), wordQueryParam)
	if err != nil {
		err = fmt.Errorf("parse query param %q in URL %s: %w", wordQueryParam, r.URL, err)
		notify(err, http.StatusBadRequest)
		return
	}

	key := keygen.New(word)
	values, err := s.Cache.Get(context.Background(), key)
	switch {
	case errors.Is(err, cache.EmptyResultErr), errors.Is(err, cache.NotValidCacheErr):
		var code int
		values, code, err = s.updateCache(key)
		if err != nil {
			err = fmt.Errorf("update cache by key %q: %v", key, err)
			notify(err, code)
			return
		}
	case err != nil:
		err = fmt.Errorf("get value from cache by key %q: %v", key, err)
		notify(err, http.StatusInternalServerError)
		return
	}

	_, _ = rw.Write(values)
}

func (s *Server) LoadHandler(rw http.ResponseWriter, r *http.Request) {
	notify := func(err error, code int) {
		log.Printf("LoadHandler: %v", err)
		rw.WriteHeader(code)
	}

	processed, code, err := processInput(r.Body)
	if err != nil {
		err = fmt.Errorf("process input: %v", err)
		notify(err, code)
		return
	}

	for key, values := range processed {
		err = s.Cache.Put(context.Background(), key, values...)
		switch {
		case errors.Is(err, cache.NotUniqWordErr):
			// TODO: think about behaviour
			log.Printf("not uniq words %v for key %q\n", values, key)
			continue
		case err != nil:
			err = fmt.Errorf("put values %q to store by key %q: %v", values, key, err)
			notify(err, http.StatusInternalServerError)
			return
		}

		object := store.Object{
			Key:   key,
			Words: values,
		}

		wordsCount, err := s.Store.Put(context.Background(), object)
		if err != nil {
			err = fmt.Errorf("put object to store key %s with values %v: %w", key, values, err)
			notify(err, http.StatusInternalServerError)
			return
		}

		log.Printf("set words count for key %s: %v\n", key, wordsCount)
		err = s.Cache.Set(context.Background(), key, wordsCount)
		if err != nil {
			err = fmt.Errorf("set object to cache: %w", err)
			notify(err, http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) updateCache(key string) ([]byte, int, error) {
	object := new(store.Object)
	err := s.Store.Get(context.Background(), key, object)
	switch {
	case errors.Is(err, store.EmptyResultErr):
		return nil, http.StatusNotFound, err
	case err != nil:
		return nil, http.StatusInternalServerError, fmt.Errorf("getting data from store: %w", err)
	}

	if len(object.Words) == 0 {
		return nil, http.StatusNotFound, fmt.Errorf("empty words")

	}

	err = s.Cache.Put(context.Background(), key, object.Words...)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("put data into cache: %w", err)
	}

	err = s.Cache.Set(context.Background(), key, len(object.Words))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("set object to cache: %w", err)
	}

	values, err := json.Marshal(object.Words)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("marshal words: %w", err)
	}

	return values, http.StatusOK, nil
}

func processInput(closer io.ReadCloser) (map[string][]string, int, error) {
	body, err := ioutil.ReadAll(closer)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("read request body: %w", err)
	}
	defer closer.Close()

	var in []string
	err = json.Unmarshal(body, &in)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("unmarshal request body: %w", err)
	}

	processed := make(map[string][]string)
	for _, v := range in {
		key := keygen.New(v)
		if _, ok := processed[key]; ok {
			processed[key] = append(processed[key], v)
			continue
		}
		processed[key] = []string{v}
	}

	return processed, http.StatusOK, nil
}

func getQueryParam(query url.Values, param string) (string, error) {
	words, ok := query[param]
	if !ok {
		return "", errors.New("there are no query param")
	}

	if len(words) != 1 {
		return "", errors.New("there are more than 1 query param")
	}

	return words[0], nil
}
