package markerMaker

type NavBar struct {
	Href    string
	Name    string
	Content string
	Active  string
}

var navs = []NavBar{
	NavBar{
		"probeset",
		"Probeset ID",
		"Search by Probeset ID",
		"Yes",
	},
	NavBar{
		"ucsc",
		"UCSC Region",
		"Search by UCSC Region",
		"No",
	},
}

var navBars = struct {
	NavBars []NavBar
}{
	navs,
}
