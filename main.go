package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/protosio/app-store/registry"

	"github.com/gorilla/mux"
	"github.com/nustiueudinastea/protos/daemon"
)

// Installer represents an application installer, but not a specific version of it.
type Installer struct {
	daemon.InstallerMetadata
	Name      string `json:"name,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

func main() {
	log.Println("Starting the Protos app store")
	// setupDB()
	mainRtr := mux.NewRouter().StrictSlash(true)
	r := mainRtr.PathPrefix("/api/v1").Subrouter()

	r.HandleFunc("/search", search).Methods("GET")
	r.HandleFunc("/event", processEvent).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))

}

// Search returns a list of installers that match the criteria
func search(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	if val, ok := queryParams["provides"]; ok {
		if len(val) == 0 {
			http.Error(w, "No value for query parameter", http.StatusInternalServerError)
		}
		installers := searchDB(val[0])
		json.NewEncoder(w).Encode(installers)
		return
	}
	http.Error(w, "'provides' is the only valid search parameter", http.StatusInternalServerError)
}

func processEvent(w http.ResponseWriter, r *http.Request) {
	bodyJSON, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}

	var events registry.Events
	err = json.Unmarshal(bodyJSON, &events)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	registry.ProcessEvents(events.Events)
}
