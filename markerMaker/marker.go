package markerMaker

// current information returned for query
// queries can be set up to handle pos<X&&chr==2, etc. This would allow getting all markers for an interval across any/all arrays
// could also do RS ID searches, etc
type Marker struct {
	MarkerName string   `json:"markerName"`
	RSId       string   `json:"rsID"`
	Chromosome string   `json:"chromosome"`
	Position   int      `json:"position"`
	A_Allele   string   `json:"a_allele"`
	B_Allele   string   `json:"b_allele"`
	Arrays     []string `json:"arrays"` //which arrays this marker is present on
}
