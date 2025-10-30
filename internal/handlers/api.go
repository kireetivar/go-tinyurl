package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/kireetivar/go-tinyurl/pkg/utils"
)

type ShortenRequest struct {
	Url string
}

type Server struct {
	InMemory map[string]ShortenRequest
	mut      sync.Mutex
}

type ShortenResponse struct {
	Hash string
}


func (s *Server) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req ShortenRequest
		decoder := json.NewDecoder(r.Body)
		err  := decoder.Decode(&req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			log.Println(err)
			return
		}
		if req.Url == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		shortKey := utils.GenerateShortKey()
		for {
			s.mut.Lock()
			_,ok := s.InMemory[shortKey]
			if !ok {
				s.InMemory[shortKey] = req
				s.mut.Unlock()
				break
			}
			s.mut.Unlock()
			shortKey = utils.GenerateShortKey()
		}

		shortres := ShortenResponse {
			Hash: shortKey,
		}

		w.Header().Set("Content-Type","application/json")
		err = json.NewEncoder(w).Encode(shortres)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		shortKey := r.PathValue("shortKey")
		fmt.Println("shortkey: " + shortKey)
		if shortKey == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		s.mut.Lock()
		req,ok := s.InMemory[shortKey]
		s.mut.Unlock()
		if !ok {
			http.Error(w, "invalid hash no url found", http.StatusNotFound)
			return
		}

		redirectURL := req.Url
		if !strings.HasPrefix(redirectURL, "http://") && !strings.HasPrefix(redirectURL, "https://") {
			redirectURL = "https://" + redirectURL
		}

		log.Printf("Redirecting %s → %s", shortKey, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusPermanentRedirect)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func NewServer() *Server {
	return &Server{
		InMemory: make(map[string]ShortenRequest),
		mut: sync.Mutex{},
	}
}