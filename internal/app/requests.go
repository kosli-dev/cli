package app

// HTTPResponse is a simplified version of http.Response
type HTTPResponse struct {
	Body       string
	StatusCode int
}

// HarvestRequest represents the POST request body to be sent to merkely harvest endpoint
type HarvestRequest struct {
	PodsData    []*PodData
	Owner       string
	Environment string
}
