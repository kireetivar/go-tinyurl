package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kireetivar/go-tinyurl/internal/models"
	"github.com/kireetivar/go-tinyurl/internal/storage"
)

func TestShortenHandler_HappyPath(t *testing.T) {
	mockStore := &storage.MockStore{
		WantSaveURL: "abcde",
		WantSaveErr: nil,
	}

	mockUserStore := &storage.MockUserStore{}

	fakeSecret := "not-real-just-for-testing"
	server := NewServer(mockStore, mockUserStore, fakeSecret)

	fakeUser := &models.User{ID: 1, Username: "test-user"}
	tokkenString, err :=  server.jwtHandler(fakeUser)
	if err!= nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	requestBody := `{"url": "https://google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(requestBody))
	req.Header.Set("Authorization", "Bearer "+tokkenString)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	server.ShortenMiddleware(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %d want %d", rr.Code, http.StatusOK)
	}

	expectedBody := `{"hash":"abcde"}`
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("handler returned unexpected body: got %q want %q", strings.TrimSpace(rr.Body.String()), expectedBody)
	}

}

func TestShortenHandler_Notoken(t *testing.T) {
	mockStore := &storage.MockStore{
		WantSaveURL: "abcde",
		WantSaveErr: nil,
	}

	mockUserStore := &storage.MockUserStore{}

	fakeSecret := "dummy-secrect"
	server := NewServer(mockStore, mockUserStore, fakeSecret)

	requestBody := `{"url" : "https://google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(requestBody))

	req.Header.Set("Authorization", "") // Dont send string so req fails
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	server.ShortenMiddleware(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}

}

func TestShortenMiddleware_InvalidToken(t *testing.T) {
	mockStore := &storage.MockStore{
		WantSaveURL: "abcde",
		WantSaveErr: nil,
	}

	mockUserStore := &storage.MockUserStore{}

	fakeSecret := "dummy-secret"
	server := NewServer(mockStore, mockUserStore, fakeSecret)

	requestBody := `{"url":"http://ww.google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(requestBody))
	req.Header.Set("Authorization","Bearer " + "faketoken")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	server.ShortenMiddleware(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %d want %d", rr.Code, http.StatusUnauthorized)
	}

}


func TestShortenMiddleware_StoreError(t *testing.T) {
	mockStore := &storage.MockStore{
		WantSaveURL: "adcde",
		WantSaveErr: errors.New("database is down"),
	}

	makeUserStore := &storage.MockUserStore{}

	fakeSecret := "dummy-secret"
	server := NewServer(mockStore, makeUserStore, fakeSecret)

	mockUser := &models.User{ID: 1, Username: "mockuser"}
	tokenString, err := server.jwtHandler(mockUser)
	if err!= nil {
		t.Errorf("Failed to generate test token: %v", err)
	}

	requestBody := `{"url":"https://google.com"}`
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(requestBody))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	server.ShortenHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %d want %d", rr.Code, http.StatusInternalServerError)
	}
}