package throttler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	log "github.com/sirupsen/logrus"
)

type ctxKey string

const (
	ctxToken ctxKey = "token"
)

// LoggingMiddleware outputs requests path and response status,
// including request-response duration
func (s *Service) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		l := log.WithFields(log.Fields{
			"environment":    s.environment,
			"request-path":   r.RequestURI,
			"request-method": r.Method,
		})
		l.Infoln()
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		m := httpsnoop.CaptureMetrics(next, w, r)
		l.WithFields(log.Fields{
			"request-duration": m.Duration,
			"response-code":    m.Code,
		}).Infoln("handler response")
	})
}

// ValidateAccessToken middleware and sets token info into context
func (s *Service) ValidateAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.writeError(w, errMissingAccessToken.msg("ValidateAccessToken.Authorization"))
			return
		}

		// TODO: Make this a bit more robust, parsing-wise
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			s.writeError(w, errMissingAccessToken.msg("ValidateAccessToken.Bearer"))
			return
		}

		token := authHeaderParts[1]
		if token == "" {
			s.writeError(w, errMissingAccessToken.msg("ValidateAccessToken.Bearer"))
			return
		}
		ctx := context.WithValue(r.Context(), ctxToken, token)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// CheckLimitsMiddlware checks the in-memory cache for token usage
// it adds an entry to the map if it's a new token.
// It checks for time passed a
func (s *Service) CheckLimitsMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxToken := r.Context().Value(ctxToken)
		if ctxToken == nil {
			s.writeError(w, errInternalServerError.msg("CheckLimitsMiddlware failed to get token from context"))
			return
		}
		token := ctxToken.(string)
		now := time.Now()
		cachedToken, ok := s.Throttler.cache[token]
		if !ok {
			s.Throttler.cache[token] = &requester{
				token:   token,
				counter: 0,
				endTime: now.Add(time.Duration(s.Throttler.M) * time.Millisecond),
			}
			cachedToken = s.Throttler.cache[token]
		}
		cachedToken.counter++

		log.Infof("cached token: %s now: %s reset-time: %s\n", cachedToken.token, now, cachedToken.endTime)

		if cachedToken.counter > s.Throttler.N {
			if now.Before(cachedToken.endTime) { // limit reached already
				timeLeft := cachedToken.endTime.Sub(now).Seconds() * 1000
				s.writeError(w, &Error{Code: http.StatusTooManyRequests, Message: fmt.Sprintf("Too many requests: %.2fms left until reset", timeLeft)})
				return
			}
			cachedToken.counter = 1
			cachedToken.endTime = now.Add(time.Duration(s.Throttler.M) * time.Millisecond)
		}

		next.ServeHTTP(w, r)

	})
}
