package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/SKevo18/server_updater/manifest"
	log "github.com/gwillem/go-simplelog"
)

var CanonicalModrinthApiUrl = getModrinthApiUrl()

const (
	apiUrlRoot        = "https://api.modrinth.com/v2"
	stagingApiUrlRoot = "https://staging-api.modrinth.com/v2"
)

func getModrinthApiUrl() string {
	if os.Getenv("MODRINTH_STAGING") == "true" {
		return stagingApiUrlRoot
	}
	return apiUrlRoot
}

// GetProject gets a project from the Modrinth API
func GetProject(projectIdOrSlug string) (*ModrinthProject, error) {
	var project ModrinthProject
	err := get(fmt.Sprintf("%s/project/%s", CanonicalModrinthApiUrl, projectIdOrSlug), &project)
	return &project, err
}

// GetVersionsFor gets all versions of a project
func GetVersionsFor(project *ModrinthProject, server manifest.Server) ([]ModrinthVersion, error) {
	var versions []ModrinthVersion

	// Properly encode query parameters
	params := url.Values{}
	params.Add("loaders", fmt.Sprintf("[\"%s\"]", server.Loader))
	params.Add("game_versions", fmt.Sprintf("[\"%s\"]", server.MinecraftVersion))

	requestUrl := fmt.Sprintf("%s/project/%s/version?%s",
		CanonicalModrinthApiUrl,
		project.ID,
		params.Encode(),
	)

	err := get(requestUrl, &versions)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// GetAllVersionsFor gets all versions of a project without compatibility filtering
func GetAllVersionsFor(project *ModrinthProject) ([]ModrinthVersion, error) {
	var versions []ModrinthVersion

	requestUrl := fmt.Sprintf("%s/project/%s/version",
		CanonicalModrinthApiUrl,
		project.ID,
	)

	err := get(requestUrl, &versions)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// GetVersion gets a specific version from the Modrinth API
func GetVersion(versionId string) (*ModrinthVersion, error) {
	var version ModrinthVersion
	err := get(fmt.Sprintf("%s/version/%s", CanonicalModrinthApiUrl, versionId), &version)
	return &version, err
}

// ResolveVersion resolves the wanted version string to a specific ModrinthVersion
func ResolveVersion(versions []ModrinthVersion, wantedVersion string) *ModrinthVersion {
	if wantedVersion == "@latest" {
		return &versions[0]
	}

	for _, v := range versions {
		if v.VersionNumber == wantedVersion {
			return &v
		}
	}
	return nil
}

// GetRequiredDependencies gets the required dependencies for a version
func GetRequiredDependencies(version *ModrinthVersion, server manifest.Server, downloadIncompatible bool) ([]*manifest.Dependency, error) {
	deps := make([]*manifest.Dependency, 0)
	for _, dep := range version.Dependencies {
		if dep.DependencyType == "required" {
			if dep.ProjectID == nil {
				log.Warn("Skipping dependency with no project ID")
				continue
			}

			depProject, err := GetProject(*dep.ProjectID)
			if err != nil {
				return nil, err
			}

			var depVersions []ModrinthVersion
			if downloadIncompatible {
				depVersions, err = GetAllVersionsFor(depProject)
			} else {
				depVersions, err = GetVersionsFor(depProject, server)
			}

			if err != nil {
				return nil, err
			}
			if len(depVersions) == 0 {
				log.Warn(fmt.Sprintf("No versions found for dependency %s", depProject.Title))
				continue
			}

			// We just take the latest compatible version for the dependency
			depVersion := ResolveVersion(depVersions, "@latest")

			var primaryFile *ModrinthFile
			for _, f := range depVersion.Files {
				if f.Primary {
					primaryFile = &f
					break
				}
			}

			if primaryFile == nil {
				log.Warn(fmt.Sprintf("No primary file found for dependency %s", depProject.Title))
				continue
			}

			newDep := &manifest.Dependency{
				ProjectId:            depProject.ID,
				Version:              depVersion.VersionNumber,
				WantedVersion:        depVersion.VersionNumber,
				FileName:             primaryFile.Filename,
				FileHash:             primaryFile.Hashes["sha512"],
				DownloadUrl:          primaryFile.URL,
				SaveAs:               primaryFile.Filename, // Dependencies are saved as their original filename
				DownloadIncompatible: downloadIncompatible, // Inherit from parent
				Metadata: map[string]any{
					"source.modrinth": map[string]any{
						"projectId": depProject.ID,
					},
				},
			}
			newDep.Dependencies, err = GetRequiredDependencies(depVersion, server, downloadIncompatible)
			if err != nil {
				return nil, err
			}
			deps = append(deps, newDep)
		}
	}
	return deps, nil
}

// DownloadFile downloads a file from a URL to a specific path
func DownloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func get(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}
