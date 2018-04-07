package rest

import (
	"encoding/json"
	"net/http"
)

// RespondWithError returns an error message formatted as an JSON object, and sets the http status to code
func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON returns a message formatted as JSON, and sets the http status to code
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
