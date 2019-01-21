package http

import (
	"net/http"
	"encoding/json"
	"gitlab.appsflyer.com/Architecture/af-go-health"
)

const (
	ReportTypeShort = "short"
)

// HandleHealthJson returns an HandlerFunc that can be used as an endpoints that exposes the service health
func HandleHealthJson(h health.Health) http.HandlerFunc {
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
		if request.URL.Query().Get("type") == ReportTypeShort {
			shortResults := make(map[string]string)
			for k, v := range results {
				if v.IsHealthy() {
					shortResults[k] = "PASS"
				} else {
					shortResults[k] = "FAIL"
				}
			}

			encoder.Encode(shortResults)
		} else {
			encoder.Encode(results)
		}

	}
}
