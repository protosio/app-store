package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/protosio/app-store/installer"
	"github.com/protosio/app-store/registry"
	"github.com/protosio/app-store/util"
)

var log = util.GetLogger()

// StartWebServer starts the webserver on the provided port
func StartWebServer(port int) {
	log.Infof("Starting the web server on port %d", port)
	mainRtr := mux.NewRouter().StrictSlash(true)
	r := mainRtr.PathPrefix("/api/v1").Subrouter()

	r.HandleFunc("/search", search).Methods("GET")
	r.HandleFunc("/event", processEvent).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))

}

func search(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	if val, ok := queryParams["provides"]; ok {
		if len(val) == 0 {
			http.Error(w, "No value for query parameter", http.StatusInternalServerError)
		}
		installers, err := installer.Search(val[0])
		if err != nil {
			log.Errorf("Can't perform search: %v", err)
			http.Error(w, "Internal error: can't perform search", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(installers)
		return
	}
	http.Error(w, "'provides' is the only valid search parameter", http.StatusInternalServerError)
}

func processEvent(w http.ResponseWriter, r *http.Request) {
	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var events registry.Events
	err = json.Unmarshal(bodyJSON, &events)
	if err != nil {
		log.Errorf("Error reading body: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	registry.ProcessEvents(events.Events)
}
