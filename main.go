package main

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/apex/log/handlers/text"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

func init() {
	if os.Getenv("UP_STAGE") == "" {
		log.SetHandler(text.Default)
	} else {
		log.SetHandler(jsonhandler.Default)
	}
}

var views = template.Must(template.New("").ParseGlob("templates/*.html"))

func main() {
	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()
	app.HandleFunc("/", index)
	app.HandleFunc("/set", set).Methods("POST")
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	views.ExecuteTemplate(w, "index.html", nil)
}

func set(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.WithError(err).Error("failed to parse form")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var selection struct {
		Due      string `schema:"due,required"`
		Timezone string `schema:"timezone,required"`
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(&selection, r.PostForm)
	if err != nil {
		log.WithError(err).Error("failed to decode")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Infof("Due: %#v", selection)

	loc, err := time.LoadLocation(selection.Timezone)
	if err != nil {
		log.WithError(err).Error("bad timezone")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	when, err := time.ParseInLocation("2006-01-02T15:04", selection.Due, loc)
	if err != nil {
		log.WithError(err).Error("bad time")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	views.ExecuteTemplate(w, "index.html", struct {
		When time.Duration
	}{
		time.Until(when),
	})
}
