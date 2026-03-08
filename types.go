package tinyserp

// SearchResponse is the JSON payload returned by the search endpoint.
type SearchResponse struct {
	SearchInformation SearchInformation `json:"searchInformation"`
	Items             []SearchItem      `json:"items"`
}

// SearchInformation describes the executed query.
type SearchInformation struct {
	Query           string `json:"query"`
	Engine          string `json:"engine"`
	ResultsReturned int    `json:"resultsReturned"`
}

// SearchItem is a single parsed search result.
type SearchItem struct {
	Rank    int    `json:"rank"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}
