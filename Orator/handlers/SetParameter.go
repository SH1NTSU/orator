package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Parameters struct {
	Rate      *int     `json:"rate,omitempty"`
	Voice     *string  `json:"voice,omitempty"`
	Volume    *float64 `json:"volume,omitempty"`
	Amplitude *int     `json:"amplitude,omitempty"`
}

var (
	speechParams = Parameters{
		Rate:      intPtr(150),
		Voice:     strPtr("default"),
		Volume:    floatPtr(1.0),
		Amplitude: intPtr(100),
	}
	mu sync.Mutex
)

func intPtr(i int) *int           { return &i }
func strPtr(s string) *string     { return &s }
func floatPtr(f float64) *float64 { return &f }

func setParameter(w http.ResponseWriter, r *http.Request) {
	var newParams Parameters
	if err := json.NewDecoder(r.Body).Decode(&newParams); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if newParams.Rate != nil {
		speechParams.Rate = newParams.Rate
	}
	if newParams.Voice != nil {
		speechParams.Voice = newParams.Voice
	}
	if newParams.Volume != nil {
		speechParams.Volume = newParams.Volume
	}
	if newParams.Amplitude != nil {
		speechParams.Amplitude = newParams.Amplitude
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(speechParams)
}
