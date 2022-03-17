package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"udemy_drone/go_tello_edu/config"
)

func viewIndexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("app/views/index.html")
	if err != nil {
		panic(err.Error())
	}
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartWebServer() error {
	http.HandleFunc("/", viewIndexHandler)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", config.Config.Address, config.Config.Port), nil)
}
