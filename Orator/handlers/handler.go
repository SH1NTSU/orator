package handlers

import (
	"net/http"
)

type Order struct{}

func (o *Order) ToSpeach(w http.ResponseWriter, r *http.Request) {
	ToSpeech(w, r)
}
func (o *Order) SetParameter(w http.ResponseWriter, r *http.Request) {
	SetParameter(w, r)
}

// func (o *Order) FetchCities(w http.ResponseWriter, r *http.Request) {
//     api.FetchCities(w, r)
// }
//
//     fmt.Println("Read handler")
// }
// func (o *Order) Delete(w http.ResponseWriter, r *http.Request) {
//
//     fmt.Println("Delete handler")
// }
// handlers rigt here
