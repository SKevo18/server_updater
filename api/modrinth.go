package api

import (
	"fmt"
	"os"
	"slices"

	"github.com/SKevo18/server_updater/manifest"
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

// FindModrinthVersionFor finds the appropriate version of a plugin/mod for a specific server type and game version
func FindModrinthVersionFor(projectIdOrSlug string, wantedVersion string, server manifest.Server) (*manifest.Dependency, error) {
	versions, err := FetchJsonArray(fmt.Sprintf("%s/project/%s/version", CanonicalModrinthApiUrl, projectIdOrSlug))
	if err != nil {
		return nil, err
	}

	for _, version := range versions {
		versionMap, ok := version.(jsonMapResponse)
		if !ok {
			return nil, fmt.Errorf("failed to convert version from Modrinth API: `%v` is `%T`, want `map[string]interface{}`", version, version)
		}

		dependency, err := getModrinthDependencyFromVersion(versionMap, wantedVersion, server)
		if err != nil {
			return nil, err
		}
		if dependency != nil {
			return dependency, nil
		}
	}

	return nil, nil
}

// getModrinthDependencyFromVersion gets the dependency from a version map,
// returns nil if the version is not found
func getModrinthDependencyFromVersion(versionMap jsonMapResponse, wantedVersion string, server manifest.Server) (*manifest.Dependency, error) {
	var dependency *manifest.Dependency

	loaders := make([]string, 0, len(versionMap["loaders"].([]any)))
	for _, rawLoader := range versionMap["loaders"].([]any) {
		loader, ok := rawLoader.(string)
		if !ok {
			return nil, fmt.Errorf("failed to convert loader from Modrinth API: `%v` is `%T`, want `string`", rawLoader, rawLoader)
		}
		loaders = append(loaders, loader)
	}

	if slices.Contains(loaders, server.Loader) {
		// loader matches, check MC version
		gameVersions := make([]string, 0, len(versionMap["game_versions"].([]any)))
		for _, rawGameVersion := range versionMap["game_versions"].([]any) {
			gameVersion, ok := rawGameVersion.(string)
			if !ok {
				return nil, fmt.Errorf("failed to convert game version from Modrinth API: `%v` is `%T`, want `string`", rawGameVersion, rawGameVersion)
			}
			gameVersions = append(gameVersions, gameVersion)
		}

		if slices.Contains(gameVersions, server.MinecraftVersion) {
			switch wantedVersion {
			case "@latest":
				// latest version is on the top of the list, so we can return it
				dependency = &manifest.Dependency{
					Version: versionMap["version_number"].(string),
				}
			case "@latestStable":
				// latest stable ("release")
				if versionMap["version_type"] == "release" {
					dependency = &manifest.Dependency{
						Version: versionMap["version_number"].(string),
					}
				}
			case versionMap["version_number"]:
				// exact match
				dependency = &manifest.Dependency{
					Version: versionMap["version_number"].(string),
				}
			default:
				return nil, nil
			}

			// find required dependencies
			requiredDependencies, err := findModrinthRequiredDependencies(versionMap["dependencies"], server)
			if err != nil {
				return nil, err
			}
			dependency.Dependencies = requiredDependencies

			// find download URL
			downloadUrl, err := getModrinthDownloadUrl(versionMap["files"])
			if err != nil {
				return nil, err
			}
			dependency.DownloadUrl = downloadUrl

			return dependency, nil
		}
	}

	return nil, nil
}

// fetchModrinthVersion fetches a raw version JSON data from the Modrinth API
func fetchModrinthVersion(versionId string) (jsonMapResponse, error) {
	versionMap, err := FetchJsonObject(fmt.Sprintf("%s/version/%s", CanonicalModrinthApiUrl, versionId))
	if err != nil {
		return nil, err
	}
	return versionMap, nil
}

// findModrinthRequiredDependencies finds the required dependencies of another dependency,
// returns nil if there are no required dependencies
func findModrinthRequiredDependencies(rawDependencies any, server manifest.Server) ([]*manifest.Dependency, error) {
	dependenciesArray, ok := rawDependencies.(jsonListResponse)
	if !ok {
		return nil, fmt.Errorf("failed to convert dependencies from Modrinth API: `%v` is `%T`, want `[]map[string]interface{}`", rawDependencies, rawDependencies)
	}

	dependencies := []*manifest.Dependency{}
	for _, rawDependency := range dependenciesArray {
		dependencyMap, ok := rawDependency.(jsonMapResponse)
		if !ok {
			return nil, fmt.Errorf("failed to convert dependency from Modrinth API: `%v` is `%T`, want `map[string]interface{}`", rawDependency, rawDependency)
		}

		if dependencyMap["dependency_type"] == "required" {
			versionId, ok := dependencyMap["version_id"].(string)
			if !ok {
				return nil, fmt.Errorf("failed to convert version id from Modrinth API: `%v` is `%T`, want `string`", dependencyMap["version_id"], dependencyMap["version_id"])
			}

			rawDependencyVersion, err := fetchModrinthVersion(versionId)
			if err != nil {
				return nil, err
			}

			dependency, err := getModrinthDependencyFromVersion(rawDependencyVersion, "@latest", server)
			if err != nil {
				return nil, err
			}

			dependencies = append(dependencies, dependency)
		}
	}

	if len(dependencies) == 0 {
		return nil, nil
	}

	return dependencies, nil
}

// getModrinthDownloadUrl obtains the download URL of a dependency from raw files data
func getModrinthDownloadUrl(rawFiles any) (string, error) {
	filesArray, ok := rawFiles.(jsonListResponse)
	if !ok {
		return "", fmt.Errorf("failed to convert files from Modrinth API: `%v` is `%T`, want `[]map[string]interface{}`", rawFiles, rawFiles)
	}

	for _, rawFile := range filesArray {
		file, ok := rawFile.(jsonMapResponse)
		if !ok {
			return "", fmt.Errorf("failed to convert file from Modrinth API: `%v` is `%T`, want `map[string]interface{}`", rawFile, rawFile)
		}

		if file["primary"] == true {
			downloadUrl, ok := file["url"].(string)
			if !ok {
				return "", fmt.Errorf("failed to convert download url from Modrinth API: `%v` is `%T`, want `string`", file["url"], file["url"])
			}
			return downloadUrl, nil
		}
	}

	return "", fmt.Errorf("no primary file found")
}
