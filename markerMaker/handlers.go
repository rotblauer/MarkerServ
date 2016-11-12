package markerMaker

import (
	"encoding/json"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

func Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	template.Must(template.New("Data").Parse(FormTemplate)).Execute(w, nil)
	template.Must(template.New("Data").Parse(PrintTemplate)).Execute(w, nil)
}

//populator, temp for adding data
func populate(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var marker Marker

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&marker)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	key := markerKey(c, marker.MarkerName)
	if _, err := datastore.Put(c, key, &marker); err != nil { //store it
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
