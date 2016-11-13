package markerMaker

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"
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

// 1	10000	500000
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

const (
	chromField = iota
	startField
	endField
)

type Bed3 struct {
	Chrom      string
	ChromStart int
	ChromEnd   int
}

func parseBed3(line string) (b *Bed3, err error) {
	const n = 3
	f := strings.Split(line, "\t")
	if len(f) < n {
		return nil, errors.New("bed: bad bed type")
	}
	b = &Bed3{
		Chrom:      string(f[chromField]),
		ChromStart: mustAtoi(f[startField], startField),
		ChromEnd:   mustAtoi(f[endField], endField),
	}
	return
}
func mustAtoi(f string, column int) int {
	i, err := strconv.ParseInt(f, 0, 0)
	if err != nil {
		panic(&csv.ParseError{Column: column, Err: err})
	}
	return int(i)
}

func (b *Bed3) Start() int { return b.ChromStart }
func (b *Bed3) End() int   { return b.ChromEnd }
func (b *Bed3) Len() int   { return b.ChromEnd - b.ChromStart }
