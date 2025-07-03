package manifest

import "strings"

type Dependency struct {
	// SaveAs name as it should be saved on disk, can contain "{version}" to be replaced with the wanted version
	SaveAs string `json:"saveAs"`

	// Wanted version, can be "@latest" to get the latest version
	Version string `json:"version"`

	// Metadata for the dependency
	Metadata map[string]any `json:"metadata"`

	// Dependencies for the dependency. nil if there are no dependencies
	Dependencies []*Dependency `json:"-"`

	// Download URL of the jar plugin/mod
	DownloadUrl string `json:"-"`
}

// CanonicalFileName returns the file name as it would be saved on the filesystem
func (d *Dependency) CanonicalFileName() string {
	return strings.ReplaceAll(d.SaveAs, "{version}", d.Version)
}

type (
	Plugin = Dependency
	Mod    = Dependency
)
