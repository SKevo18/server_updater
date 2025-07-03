package api

import "time"

type HangarProject struct {
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Namespace   Namespace `json:"namespace"`
	Stats       Stats     `json:"stats"`
	Category    string    `json:"category"`
	LastUpdated time.Time `json:"lastUpdated"`
	Visibility  string    `json:"visibility"`
	AvatarURL   string    `json:"avatarUrl"`
	Description string    `json:"description"`
	UserActions struct {
		Starred  bool `json:"starred"`
		Watching bool `json:"watching"`
		Flagged  bool `json:"flagged"`
	} `json:"userActions"`
	Settings struct {
		Links     []Link   `json:"links"`
		Tags      []string `json:"tags"`
		License   License  `json:"license"`
		Keywords  []string `json:"keywords"`
		Sponsors  string   `json:"sponsors"`
		Donation  Donation `json:"donation"`
		Homepage  string   `json:"homepage"`
		Issues    string   `json:"issues"`
		Source    string   `json:"source"`
		Support   string   `json:"support"`
		Wiki      string   `json:"wiki"`
		ForumSync bool     `json:"forumSync"`
	} `json:"settings"`
	ProjectID int `json:"projectId"`
}

type Namespace struct {
	Owner string `json:"owner"`
	Slug  string `json:"slug"`
}

type Stats struct {
	Views           int `json:"views"`
	Downloads       int `json:"downloads"`
	RecentViews     int `json:"recentViews"`
	RecentDownloads int `json:"recentDownloads"`
	Stars           int `json:"stars"`
	Watchers        int `json:"watchers"`
}

type Link struct {
	ID    int        `json:"id"`
	Type  string     `json:"type"`
	Title string     `json:"title"`
	Links []LinkItem `json:"links"`
}

type LinkItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Donation struct {
	Enable  bool   `json:"enable"`
	Subject string `json:"subject"`
}

type HangarVersionsResponse struct {
	Pagination Pagination      `json:"pagination"`
	Result     []HangarVersion `json:"result"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type HangarVersion struct {
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Visibility  string    `json:"visibility"`
	Description string    `json:"description"`
	Stats       Stats     `json:"stats"`
	Author      string    `json:"author"`
	ReviewState string    `json:"reviewState"`
	Channel     Channel   `json:"channel"`
	PinnedUsers []string  `json:"pinnedUsers"`
	Downloads   struct {
		// Platform-specific downloads (e.g., "PAPER", "VELOCITY", etc.)
	} `json:"downloads"`
	PlatformDependencies          map[string][]string `json:"platformDependencies"`
	PlatformDependenciesFormatted map[string][]string `json:"platformDependenciesFormatted"`
}

type Channel struct {
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	Flags       []string  `json:"flags"`
}

type PlatformDependency struct {
	Name             string   `json:"name"`
	Required         bool     `json:"required"`
	PlatformVersions []string `json:"platformVersions"`
}
