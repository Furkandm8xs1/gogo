package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type Job struct {
	ID        int       `json:"id"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type Server struct {
	jobs chan Job

	mu        sync.RWMutex
	received  int
	processed int
}

func NewServer(queueSize int) *Server {
	return &Server{jobs: make(chan Job, queueSize)}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /stats", s.stats)
	mux.HandleFunc("POST /jobs", s.createJob)
	return loggingMiddleware(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) stats(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]int{
		"received":  s.received,
		"processed": s.processed,
		"queued":    len(s.jobs),
	})
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Payload == "" {
		http.Error(w, "body must contain a non-empty payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.received++
	job := Job{
		ID:        s.received,
		Payload:   input.Payload,
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Unlock()

	// Handler request context'i kapanırsa enqueue işlemi de beklemeden iptal olur.
	select {
	case s.jobs <- job:
		writeJSON(w, http.StatusAccepted, job)
	case <-r.Context().Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
	}
}

func (s *Server) worker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("worker=%d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("worker=%d stopped: %v", id, ctx.Err())
			return
		case job := <-s.jobs:
			log.Printf("worker=%d processing job=%d payload=%q", id, job.ID, job.Payload)

			// Gerçek hayatta burada API çağrısı, mesaj gönderimi veya DB işlemi olabilir.
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				log.Printf("worker=%d cancelled job=%d", id, job.ID)
				return
			}

			s.mu.Lock()
			s.processed++
			s.mu.Unlock()
			log.Printf("worker=%d completed job=%d", id, job.ID)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s duration=%s", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func main() {
	port := envInt("PORT", 8080)
	workerCount := envInt("WORKERS", 3)
	queueSize := envInt("QUEUE_SIZE", 10)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := NewServer(queueSize)
	var workers sync.WaitGroup
	for id := 1; id <= workerCount; id++ {
		workers.Add(1)
		go server.worker(ctx, id, &workers)
	}

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		Handler:      server.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%d workers=%d queue=%d", port, workerCount, queueSize)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	workers.Wait()
	log.Println("shutdown completed")
}
