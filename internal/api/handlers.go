package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h apiHandler) handleEcho(w http.ResponseWriter, r * http.Request){
	message := chi.URLParam(r, "message")
	fmt.Println(r.URL)
	fmt.Println("Hello", message)
	w.Write([]byte("echo " + message))
}