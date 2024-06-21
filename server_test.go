package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestRequest(method, targetURL, body string) *http.Request {
	req := httptest.NewRequest(method, "/?url="+url.QueryEscape(targetURL), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func checkCorsHeaders(t *testing.T, w *httptest.ResponseRecorder) {
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header Access-Control-Allow-Origin is missing or incorrect")
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Error("CORS header Access-Control-Allow-Methods is missing or incorrect")
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Error("CORS header Access-Control-Allow-Headers is missing or incorrect")
	}
}

func TestProxyHandler(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			w.Write(body)
		} else {
			w.Write([]byte("backend response"))
		}
	}))
	defer backend.Close()

	t.Run("GET request", func(t *testing.T) {
		req := newTestRequest(http.MethodGet, backend.URL, "")
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "backend response" {
			t.Errorf("Expected 'backend response', got '%s'", string(body))
		}
		checkCorsHeaders(t, w)
	})

	t.Run("POST request", func(t *testing.T) {
		req := newTestRequest(http.MethodPost, backend.URL, "test body")
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		if string(body) != "test body" {
			t.Errorf("Expected 'test body', got '%s'", string(body))
		}
		checkCorsHeaders(t, w)
	})

	t.Run("Missing URL parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Missing url parameter\n" {
			t.Errorf("Expected 'Missing url parameter', got '%s'", string(body))
		}
		checkCorsHeaders(t, w)
	})

	t.Run("Invalid URL parameter", func(t *testing.T) {
		req := newTestRequest(http.MethodGet, "http://%", "")
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Invalid url parameter\n" {
			t.Errorf("Expected 'Invalid url parameter', got '%s'", string(body))
		}
		checkCorsHeaders(t, w)
	})

	t.Run("Valid URL with JSON response", func(t *testing.T) {
		req := newTestRequest(http.MethodGet, "https://jsonplaceholder.typicode.com/posts/1", "")
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		expectedBody := `{
  "userId": 1,
  "id": 1,
  "title": "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
  "body": "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto"
}`

		if string(body) != expectedBody {
			t.Errorf("Expected JSON body, got '%s'", string(body))
		}
		if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", resp.Header.Get("Content-Type"))
		}
		checkCorsHeaders(t, w)
	})

	t.Run("POST to valid URL with JSON response", func(t *testing.T) {
		reqBody := `{
		  "title": "foo",
		  "body": "bar",
		  "userId": 1
		}`
		req := newTestRequest(http.MethodPost, "https://jsonplaceholder.typicode.com/posts", reqBody)
		w := httptest.NewRecorder()

		proxyHandler(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		// Check if the response body contains the expected fields
		expectedFields := []string{"title", "body", "userId", "id"}
		for _, field := range expectedFields {
			if !strings.Contains(string(body), field) {
				t.Errorf("Expected JSON body to contain field '%s', got '%s'", field, string(body))
			}
		}

		if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("Expected content type 'application/json; charset=utf-8', got '%s'", resp.Header.Get("Content-Type"))
		}

		checkCorsHeaders(t, w)
	})
}
