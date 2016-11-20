package markerMaker

//Handles the searching of the datastore by keys and such

import (
	"errors"
	"strings"

	"appengine"
	"appengine/datastore"
)

type queryType int

// Different types of searches supported
const (
	MARKER_NAME QueryType = 1 + iota
	UCSC_REGION
	RS_ID
)

// multi-type query organizer
func queryAll(queries []string, qType queryType, c appengine.Context) ([]Marker, []error) {

	var markers []Marker
	var errorsFound []error
	for _, query := range queries {
		var tmp []Marker
		switch qType {
		case MARKER_NAME:
			tmp = queryByName(query, c)
		case UCSC_REGION:
			bed, err := parseBed3(strings.TrimSpace(query))
			if err != nil {
				errorsFound = append(errorsFound, errors.New(query+" gave error: "+err.Error()))
			} else {
				tmp = queryByPosition(bed.Chrom, bed.Start(), bed.End(), c)
			}
		case RS_ID:
			tmp = queryByRsId(query, c)
		default:
			errorsFound = append(errorsFound, errors.New("Invalid search type"))
		}
		if len(tmp) == 0 {
			// err := errors.New("No matches for " + query + " were found")
			errorsFound = append(errorsFound, errors.New("No matches for "+query+" were found"))
		} else {
			for _, marker := range tmp {
				markers = append(markers, marker)
			}
		}
	}
	return markers, errorsFound
}

//load the info for a particular marker by marker name
func queryByName(markerName string, c appengine.Context) []Marker {

	var markers []Marker
	if strings.TrimSpace(markerName) != "" {
		marker := Marker{}
		err := datastore.Get(c, markerKey(c, markerName), &marker)
		if err == nil {
			markers = append(markers, marker)
		}
	}
	return markers
}

// set up an rsid search
func queryByRsId(rsId string, c appengine.Context) []Marker {

	q := datastore.NewQuery("Markers").
		Filter("RSId =", rsId)
	return runQuery(*q, c)
}

// set up a UCSC region query
func queryByPosition(chr string, start int, stop int, c appengine.Context) []Marker {

	// The Query type and its methods are used to construct a query.
	q := datastore.NewQuery("Markers").
		Filter("Chromosome =", chr).
		Filter("Position <=", stop).
		Filter("Position >=", start)

	return runQuery(*q, c)
}

// returns all markers associated with query
func runQuery(q datastore.Query, c appengine.Context) []Marker {

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

func storeMarker(marker Marker, c appengine.Context) error {

	key := markerKey(c, marker.MarkerName)
	if _, err := datastore.Put(c, key, &marker); err != nil { //store it
		return err
	}

	return nil
}

// forms the marker key
func markerKey(c appengine.Context, markerName string) *datastore.Key {
	return datastore.NewKey(c, "Markers", strings.TrimSpace(markerName), 0, nil)
}
