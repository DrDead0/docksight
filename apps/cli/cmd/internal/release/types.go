package release

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type GithubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}