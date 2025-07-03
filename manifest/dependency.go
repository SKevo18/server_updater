package manifest

import "strings"

type Dependency struct {
	// SaveAs name as it should be saved on disk, can contain "{version}" to be replaced with the wanted version
	SaveAs string `json:"saveAs"`

	// Wanted version, can be "@latest" to get the latest version
	WantedVersion string `json:"version"`

	// Download even if MC version or loader doesn't match
	DownloadIncompatible bool `json:"downloadIncompatible"`

	// ProjectId of the dependency, used for Modrinth API calls
	ProjectId string `json:"-"`

	// Actual version downloaded from the Modrinth API
	Version string `json:"-"`

	// File name of the downloaded file
	FileName string `json:"-"`

	// Hash of the downloaded file
	FileHash string `json:"-"`

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
