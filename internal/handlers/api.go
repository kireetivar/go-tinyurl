package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kireetivar/go-tinyurl/internal/models"
	"github.com/kireetivar/go-tinyurl/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type ShortenRequest struct {
	Url string
}

type Server struct {
	store storage.Storage
	userStore storage.UserStorage
	jwtSecret string
}

type ShortenResponse struct {
	Hash string `json:"hash"`
}

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupResponse struct {
	UserId int64 `json:"userId"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
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

func (s *Server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req SignupRequest

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "missing fields in signup request", http.StatusBadRequest)
			return
		}
		hashedPassword, err:=bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err!=nil {
			log.Printf("error while hashing the password: %v\n",err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		userId,err :=s.userStore.Create(req.Username,req.Email,string(hashedPassword))
		if err != nil {
			log.Printf("error while creating user in db: %v\n",err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		res := SignupResponse{
			UserId: userId,
		}

		w.Header().Set("Content-Type","application/json")
		err = json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Printf("Error while forming response for user: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var loginRequest LoginRequest
		err := json.NewDecoder(r.Body).Decode(&loginRequest)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if loginRequest.Username == "" || loginRequest.Password == "" {
			http.Error(w, "missing fields in login request", http.StatusBadRequest)
			return
		}

		user,err := s.userStore.GetByUsername(loginRequest.Username)
		if err!=nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(loginRequest.Password))
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		tokenString, err:= s.jwtHandler(user)
		if err != nil {
			log.Printf("ERROR while generating jwt for user %s: %v\n", user.Username, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
			return
		}
		res := LoginResponse{
			Token: tokenString,
		}

		w.Header().Set("Content-Type","application/json")
		err = json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Printf("Error while forming response for user: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (s *Server) jwtHandler(user *models.User) (string,error) {
	claims := jwt.MapClaims {
		"userID": user.ID,
		"username": user.Username,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "",err
	}

	return tokenString,nil
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /signup", s.SignupHandler)
	mux.HandleFunc("POST /login", s.LoginHandler)
	mux.HandleFunc("POST /shorten", s.ShortenMiddleware) // The protected route
	mux.HandleFunc("GET /{shortKey}", s.RedirectHandler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

func NewServer(s storage.Storage, us storage.UserStorage, jwtSecret string) *Server {
	return &Server{
		store: s,
		userStore: us,
		jwtSecret: jwtSecret,
	}
}