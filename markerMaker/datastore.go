package markerMaker

//Handles the searching of the datastore by keys and such

import (
	"net/http"
	"strings"

	"appengine"
	"appengine/datastore"
)

//load the info for a particular marker by marker name
func queryByNames(markerNames []string, c appengine.Context) []Marker {
	var markers []Marker
	for _, markerName := range markerNames {
		if strings.TrimSpace(markerName) != "" {

			marker := Marker{}
			err := datastore.Get(c, markerKey(c, markerName), &marker)
			if err != nil {
				marker.MarkerName = markerName + "(" + err.Error() + ")"
			}
			markers = append(markers, marker)
		}
	}
	return markers
}

// parse the request, return all results
func queryMarker(w http.ResponseWriter, r *http.Request) []Marker {
	c := appengine.NewContext(r)
	parts := strings.Split(r.URL.Path, "/")
	id := strings.Split(parts[2], ",")
	return queryByNames(id, c)

}

func queryByPosition(chr string, start int, stop int, c appengine.Context) []Marker {

	// The Query type and its methods are used to construct a query.
	q := datastore.NewQuery("Markers").
		Filter("Chromosome =", chr).
		Filter("Position <=", stop).
		Filter("Position >=", start)

	var markers []Marker
	t := q.Run(c)
	for {
		var marker Marker
		_, err := t.Next(&marker)
		if err == datastore.Done {
			break
		}
		if err != nil {
			break
		}
		markers = append(markers, marker)
	}
	return markers
}

// forms the marker key
func markerKey(c appengine.Context, markerName string) *datastore.Key {
	return datastore.NewKey(c, "Markers", strings.TrimSpace(markerName), 0, nil)
}

// if strings.TrimSpace(r.FormValue("list")) != "" {
// 	var markers []Marker
// 	switch formType := r.FormValue("type"); formType {
