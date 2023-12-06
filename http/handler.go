package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

const (
	// ReportTypeShort is the value to be passed in the request parameter `type` when a short response is desired.
	ReportTypeShort = "short"
)

// HandleHealthJSON returns an HandlerFunc that can be used as an endpoints that exposes the service health
func HandleHealthJSON(h gosundheit.Health) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		results, healthy := h.Results()
		w.Header().Set("Content-Type", "application/json")
		if healthy {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(503)
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "\t")
		var err error
		if request.URL.Query().Get("type") == ReportTypeShort {
			shortResults := make(map[string]string)
			for k, v := range results {
				if v.IsHealthy() {
					shortResults[k] = "PASS"
				} else {
					shortResults[k] = "FAIL"
				}
			}

			err = encoder.Encode(shortResults)
		} else {
			err = encoder.Encode(results)
		}

		if err != nil {
			_, _ = fmt.Fprintf(w, "Failed to render results JSON: %s", err)
		}
	}
}
