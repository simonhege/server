package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

// ReadRequestJSON expects req to have a JSON content type with a body that
// contains a JSON-encoded value complying with the underlying type of target.
// It populates target, or returns an error.
func ReadRequestJSON(req *http.Request, target any) error {
	contentType := req.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return err
	}
	if mediaType != "application/json" {
		return fmt.Errorf("expect application/json Content-Type, got %s", mediaType)
	}

	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(target)
}

// RenderJSON renders 'v' as JSON and writes it as a response into w.
func RenderJSON(w http.ResponseWriter, v any) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// RenderJSON renders 'v' as JSON and writes it as a response into w.
func RenderPrettyJSON(w http.ResponseWriter, v any) {
	js, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
