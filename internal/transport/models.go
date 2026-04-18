package transport

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL  string `json:"short_url"`
	ShortCode string `json:"short_code"`
}
