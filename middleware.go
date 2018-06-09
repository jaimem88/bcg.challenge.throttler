package throttler

import (
	"context"
	"net/http"
	"strings"

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

// ThrottlerMiddlware ..
func (s *Service) ThrottlerMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// token := r.Context().Value(ctxToken)
		next.ServeHTTP(w, r)

	})
}
