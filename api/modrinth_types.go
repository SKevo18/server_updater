package api

type ModrinthProject struct {
	Slug        string `json:"slug"`
	ID          string `json:"id"`
	ProjectType string `json:"project_type"`
	Title       string `json:"title"`
}

type ModrinthVersion struct {
	ID            string               `json:"id"`
	ProjectID     string               `json:"project_id"`
	Name          string               `json:"name"`
	VersionNumber string               `json:"version_number"`
	Dependencies  []ModrinthDependency `json:"dependencies"`
	GameVersions  []string             `json:"game_versions"`
	VersionType   string               `json:"version_type"`
	Loaders       []string             `json:"loaders"`
	Featured      bool                 `json:"featured"`
	Files         []ModrinthFile       `json:"files"`
}

type ModrinthFile struct {
	Hashes   map[string]string `json:"hashes"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Primary  bool              `json:"primary"`
	Size     int               `json:"size"`
}

type ModrinthDependency struct {
	VersionID      *string `json:"version_id"`
	ProjectID      *string `json:"project_id"`
	DependencyType string  `json:"dependency_type"`
}
