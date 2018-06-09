package throttler

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// HandleHealthcheck responds with an empty JSON object
func (s *Service) HandleHealthcheck(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, struct{}{})
}

// HandleNotFound writes error and logs the requested method
func (s *Service) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	s.writeError(w, &Error{Code: http.StatusNotFound, Message: r.RequestURI + " not found"})
}

// HandleGetUsers GET /users
func (s *Service) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(ctxToken)
	log.Infof("HandleGetUsers for token: |%s|", token)

	s.writeJSON(w, []struct{}{})
}

func (s *Service) writeError(w http.ResponseWriter, e *Error) {

	log.WithError(e).WithField("error-message", e.message).Error()

	js, _ := json.Marshal(e)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	w.Write(js)
}

func (s *Service) writeJSON(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	js, err := json.Marshal(i)
	if err != nil {
		s.writeError(w, errInternalServerError.msg("json.Marshal: "+err.Error()))
		return
	}
	w.Write(js)
}
