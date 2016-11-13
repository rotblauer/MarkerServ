package markerMaker

//Handles the searching of the datastore by keys and such

import (
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
