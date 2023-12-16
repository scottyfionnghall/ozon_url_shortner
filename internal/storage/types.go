package storage

type URL struct {
	ID           int64  `json:"id"`
	Domain       string `json:"domain"`
	OriginalPath string `json:"original_path"`
	ShortenPath  string `json:"shorten_path"`
}

type Requset struct {
	URL string `json:"url"`
}
