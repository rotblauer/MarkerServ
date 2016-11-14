package markerMaker

import (
	"encoding/json"
	"net/http"
)

//start the url handlers
func init() {
	router := NewRouter()
	http.Handle("/", router)
}

// for a json like response - curl friendly
func queryRaw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	markers := queryMarker(w, r)
	markerJSON, err := json.Marshal(markers) // return bytes, err
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(markerJSON)
	}
}
