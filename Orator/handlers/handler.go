package handlers

import (
	"net/http"
)

type Order struct{}

func (o *Order) ToSpeach(w http.ResponseWriter, r *http.Request) {
	ToSpeech(w, r)
}
func (o *Order) SetParameter(w http.ResponseWriter, r *http.Request) {
	setParameter(w, r)
}
