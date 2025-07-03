package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/SKevo18/server_updater/manifest"
	log "github.com/gwillem/go-simplelog"
)

const HangarApiUrl = "https://hangar.papermc.io/api/v1"

// GetHangarProject gets a project from the Hangar API
func GetHangarProject(projectSlugOrId string) (*HangarProject, error) {
	var project HangarProject
	err := get(fmt.Sprintf("%s/projects/%s", HangarApiUrl, projectSlugOrId), &project)
	return &project, err
}

// GetHangarVersionsFor gets all versions of a project with compatibility filtering
func GetHangarVersionsFor(project *HangarProject, server manifest.Server) ([]HangarVersion, error) {
	params := url.Values{}

	// Map common loaders to Hangar platform names
	platform := mapLoaderToPlatform(server.Loader)
	if platform != "" {
		params.Add("platform", platform)
		params.Add("platformVersion", server.MinecraftVersion)
	}

	// Add pagination - get first 25 versions
	params.Add("limit", "25")
	params.Add("offset", "0")

	requestUrl := fmt.Sprintf("%s/projects/%s/versions?%s",
		HangarApiUrl,
		project.Namespace.Slug,
		params.Encode(),
	)

	log.Debug(fmt.Sprintf("Requesting Hangar versions: %s", requestUrl))

	var response HangarVersionsResponse
	err := get(requestUrl, &response)
	if err != nil {
		log.Debug(fmt.Sprintf("Error getting Hangar versions: %v", err))
		return nil, err
	}

	log.Debug(fmt.Sprintf("Found %d Hangar versions", len(response.Result)))

	return response.Result, nil
}

// GetAllHangarVersionsFor gets all versions of a project without compatibility filtering
func GetAllHangarVersionsFor(project *HangarProject) ([]HangarVersion, error) {
	params := url.Values{}
	params.Add("limit", "25")
	params.Add("offset", "0")

	requestUrl := fmt.Sprintf("%s/projects/%s/versions?%s",
		HangarApiUrl,
		project.Namespace.Slug,
		params.Encode(),
	)

	var response HangarVersionsResponse
	err := get(requestUrl, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

// ResolveHangarVersion resolves the wanted version string to a specific HangarVersion
func ResolveHangarVersion(versions []HangarVersion, wantedVersion string) *HangarVersion {
	if wantedVersion == "@latest" && len(versions) > 0 {
		return &versions[0]
	}

	for _, v := range versions {
		if v.Name == wantedVersion {
			return &v
		}
	}
	return nil
}

// GetHangarDownloadUrl gets the download URL for a specific platform
func GetHangarDownloadUrl(project *HangarProject, version *HangarVersion, server manifest.Server) (string, string, error) {
	platform := mapLoaderToPlatform(server.Loader)
	if platform == "" {
		return "", "", fmt.Errorf("unsupported loader: %s", server.Loader)
	}

	// The download URL follows Hangar's pattern
	downloadUrl := fmt.Sprintf("%s/projects/%s/versions/%s/%s/download",
		HangarApiUrl,
		project.Namespace.Slug,
		version.Name,
		platform,
	)

	// Generate filename based on project name and version
	filename := fmt.Sprintf("%s-%s.jar", project.Name, version.Name)
	// Clean filename of any invalid characters
	filename = strings.ReplaceAll(filename, " ", "-")

	return downloadUrl, filename, nil
}

// GetHangarRequiredDependencies gets the required dependencies for a version
func GetHangarRequiredDependencies(project *HangarProject, version *HangarVersion, server manifest.Server, downloadIncompatible bool) ([]*manifest.Dependency, error) {
	deps := make([]*manifest.Dependency, 0)

	// Currently, Hangar API doesn't provide dependency project information in a format
	// that allows automatic resolution like Modrinth does.
	// The platformDependencies field contains supported Minecraft versions, not plugin dependencies.
	// Future enhancement could parse description or other fields for dependency information.

	return deps, nil
}

// mapLoaderToPlatform maps common loader names to Hangar platform names
func mapLoaderToPlatform(loader string) string {
	switch strings.ToLower(loader) {
	case "paper", "spigot", "bukkit":
		return "PAPER"
	case "velocity":
		return "VELOCITY"
	case "waterfall", "bungeecord":
		return "WATERFALL"
	default:
		return ""
	}
}
