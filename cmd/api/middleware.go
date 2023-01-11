package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})

}

func (app *Application) rateLimiter(next http.Handler) http.Handler {
	type Clients struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu     sync.Mutex
		client = make(map[string]*Clients)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range client {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(client, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()

		if _, found := client[ip]; !found {
			client[ip] = &Clients{
				limiter: rate.NewLimiter(rate.Limit(app.Config.Limiter.rps), app.Config.Limiter.bust),
			}
		}

		client[ip].lastSeen = time.Now()

		if !client[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
