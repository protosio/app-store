package main

import (
	"encoding/json"
	"log"
	"net/http"

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
	setupDB()
	mainRtr := mux.NewRouter().StrictSlash(true)
	r := mainRtr.PathPrefix("/api/v1").Subrouter()

	r.HandleFunc("/search", Search).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))
}

// Search returns a list of installers that match the criteria
func Search(w http.ResponseWriter, r *http.Request) {
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
