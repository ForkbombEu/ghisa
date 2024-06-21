package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	proxyURL, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Invalid url parameter", http.StatusBadRequest)
		return
	}

	var req *http.Request
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "GHISA: Failed to read request body", http.StatusInternalServerError)
			return
		}
		req, err = http.NewRequest(http.MethodPost, proxyURL.String(), strings.NewReader(string(body)))
		if err != nil {
			http.Error(w, "GHISA: Failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
	} else {
		req, err = http.NewRequest(http.MethodGet, proxyURL.String(), nil)
		if err != nil {
			http.Error(w, "GHISA: Failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "GHISA: Failed to make request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "GHISA: Failed to read response body", http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.URL.Path != "/health" {
		http.NotFound(w, req)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server is healthy")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", proxyHandler)
	mux.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:              ":5552",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("GHISA: Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("GHISA: Could not gracefully shut down server: %v\n", err)
		}
		log.Println("GHISA server stopped")
	}()

	log.Println("GHISA: proxy server is running on port 5552...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("GHISA: Could not listen on port 5552: %v\n", err)
	}
}
