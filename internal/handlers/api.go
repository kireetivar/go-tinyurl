package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/kireetivar/go-tinyurl/internal/storage"
)

type ShortenRequest struct {
	Url string
}

type Server struct {
	store storage.Storage
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
		shortKey, err := s.store.Save(req.Url)
		if err != nil {
			log.Printf("error while saving to db: %v\n",err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
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
		if shortKey == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		redirectURL,err := s.store.Get(shortKey)
		if err != nil {
			http.Error(w, "invalid hash no url found", http.StatusNotFound)
			return
		}

		if !strings.HasPrefix(redirectURL, "http://") && !strings.HasPrefix(redirectURL, "https://") {
			redirectURL = "https://" + redirectURL
		}

		log.Printf("Redirecting %s → %s", shortKey, redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusPermanentRedirect)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func NewServer(s storage.Storage) *Server {
	return &Server{
		store: s,
	}
}