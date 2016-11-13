package markerMaker

type NavBar struct {
	Href    string
	Name    string
	Content string
	Active  string
	Submit  string
}

var navs = []NavBar{
	NavBar{
		"probeset",
		"Probeset ID",
		"Search by Probeset ID",
		"Yes",
		"/probesetq",
	},
	NavBar{
		"ucsc",
		"UCSC Region",
		"Search by UCSC Region",
		"No",
		"/ucscq",
	},
	NavBar{
		"rsid",
		"RS ID",
		"Search by rs ID",
		"No",
		"/rsidq",
	},
}

var navBars = struct {
	NavBars []NavBar
}{
	navs,
}
