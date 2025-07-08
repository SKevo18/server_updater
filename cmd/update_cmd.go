package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SKevo18/server_updater/api"
	"github.com/SKevo18/server_updater/manifest"
	log "github.com/gwillem/go-simplelog"
	"github.com/jlaffaye/ftp"
	"github.com/spf13/cobra"
)

var configFilePath string

const cacheFileName = "updater_cache.json"

func init() {
	updateCmd.Flags().StringVarP(&configFilePath, "config", "c", "server_manifest.json", "Path to a manifest file")
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [root_path]",
	Short: "Updates server jar and plugins defined in manifest file. If no plugins or server jar are present, downloads them.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var rootDir string
		if len(args) == 0 {
			rootDir = "."
		} else {
			rootDir = args[0]
		}

		// Read the manifest
		log.Task("Reading manifest file...")

		// Read the file directly instead of using Viper's Unmarshal
		data, err := os.ReadFile(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		var m manifest.Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
		log.Task("Manifest parsed successfully")

		// FTP client
		var ftpClient *ftp.ServerConn
		if m.FTP != nil {
			log.Task(fmt.Sprintf("Connecting to FTP server at %s", m.FTP.Host))
			var err error
			ftpClient, err = ftp.Dial(fmt.Sprintf("%s:%d", m.FTP.Host, m.FTP.Port))
			if err != nil {
				return err
			}
			defer ftpClient.Quit()

			if err := ftpClient.Login(m.FTP.Username, m.FTP.Password); err != nil {
				return err
			}
			log.Task("FTP login successful")

			// Change to remote directory if specified
			if m.FTP.RemotePath != "" {
				if err := ftpClient.ChangeDir(m.FTP.RemotePath); err != nil {
					// Try to create the directory if it doesn't exist
					if err := ftpClient.MakeDir(m.FTP.RemotePath); err != nil {
						return fmt.Errorf("failed to create remote directory %s: %w", m.FTP.RemotePath, err)
					}
					if err := ftpClient.ChangeDir(m.FTP.RemotePath); err != nil {
						return fmt.Errorf("failed to change to created remote directory %s: %w", m.FTP.RemotePath, err)
					}
				}
				log.Task(fmt.Sprintf("Changed to remote directory: %s", m.FTP.RemotePath))
			}
		}

		// Read cache
		cache := readCache(rootDir, ftpClient)
		newCache := make(map[string]string)

		// Build a map of all project IDs defined in the manifest for dependency precedence
		manifestProjectIds := buildManifestProjectIdMap(&m)

		// Process plugins
		if m.HasPlugins() {
			log.Task("Processing plugins")
			plugins := make([]*manifest.Dependency, len(m.Plugins))
			for i := range m.Plugins {
				plugins[i] = &m.Plugins[i]
			}
			err := processDependencies(plugins, "plugins", &m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			if err != nil {
				return err
			}
		}

		// Process mods
		if m.HasMods() {
			log.Task("Processing mods")
			mods := make([]*manifest.Dependency, len(m.Mods))
			for i := range m.Mods {
				mods[i] = &m.Mods[i]
			}
			err := processDependencies(mods, "mods", &m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			if err != nil {
				return err
			}
		}

		// Write new cache
		return writeCache(rootDir, ftpClient, newCache)
	},
}

// buildManifestProjectIdMap creates a map of all project IDs defined in the manifest
func buildManifestProjectIdMap(m *manifest.Manifest) map[string]bool {
	projectIds := make(map[string]bool)

	// Add all plugin project IDs
	for _, plugin := range m.Plugins {
		if projectId := extractProjectId(&plugin); projectId != "" {
			projectIds[projectId] = true
		}
	}

	// Add all mod project IDs
	for _, mod := range m.Mods {
		if projectId := extractProjectId(&mod); projectId != "" {
			projectIds[projectId] = true
		}
	}

	return projectIds
}

// extractProjectId extracts the project ID from a dependency's metadata
func extractProjectId(dep *manifest.Dependency) string {
	for sourceKey, sourceMeta := range dep.Metadata {
		if !strings.HasPrefix(sourceKey, "source.") {
			continue
		}

		sourceType := strings.TrimPrefix(sourceKey, "source.")
		switch sourceType {
		case "modrinth", "hangar":
			if meta, ok := sourceMeta.(map[string]any); ok {
				for key, value := range meta {
					if strings.EqualFold(key, "projectId") || strings.EqualFold(key, "projectSlug") {
						if projectId, ok := value.(string); ok {
							return projectId
						}
					}
				}
			}
		}
	}
	return ""
}

func processDependencies(deps []*manifest.Dependency, depType string, m *manifest.Manifest, rootDir string, ftpClient *ftp.ServerConn, cache, newCache map[string]string, manifestProjectIds map[string]bool) error {
	for _, dep := range deps {
		var sourceFound bool
		for sourceKey, sourceMeta := range dep.Metadata {
			if !strings.HasPrefix(sourceKey, "source.") {
				continue
			}
			sourceFound = true

			sourceType := strings.TrimPrefix(sourceKey, "source.")
			var err error
			switch sourceType {
			case "modrinth":
				err = processModrinthDependency(dep, sourceMeta, depType, m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			case "hangar":
				err = processHangarDependency(dep, sourceMeta, depType, m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			default:
				log.Warn(fmt.Sprintf("Unknown dependency source: %s", sourceType))
			}

			if err != nil {
				return err
			}
			// Found and processed a source, so we can break from this inner loop.
			break
		}
		if !sourceFound {
			log.Warn(fmt.Sprintf("No source found for dependency with SaveAs: %s", dep.SaveAs))
		}
	}
	return nil
}

func processModrinthDependency(dep *manifest.Dependency, meta any, depType string, m *manifest.Manifest, rootDir string, ftpClient *ftp.ServerConn, cache, newCache map[string]string, manifestProjectIds map[string]bool) error {
	modrinthMeta, ok := meta.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid modrinth metadata format for %s", dep.SaveAs)
	}

	var projectId string
	var projectIdFound bool
	for key, value := range modrinthMeta {
		if strings.EqualFold(key, "projectId") {
			projectId, projectIdFound = value.(string)
			break
		}
	}

	if !projectIdFound {
		return fmt.Errorf("projectId not found or not a string in modrinth metadata for %s", dep.SaveAs)
	}

	log.Task(fmt.Sprintf("Processing %s: %s", depType, projectId))

	project, err := api.GetProject(projectId)
	if err != nil {
		return err
	}

	var versions []api.ModrinthVersion
	if dep.DownloadIncompatible {
		versions, err = api.GetAllVersionsFor(project)
	} else {
		versions, err = api.GetVersionsFor(project, m.Server)
	}

	if err != nil || len(versions) == 0 {
		log.Warn(fmt.Sprintf("No compatible versions found for %s", dep.SaveAs))
		return nil // Continue with next dependency
	}

	version := api.ResolveVersion(versions, dep.WantedVersion)
	if version == nil {
		log.Warn(fmt.Sprintf("Wanted version '%s' not found for %s", dep.WantedVersion, dep.SaveAs))
		return nil // Continue with next dependency
	}

	// Main dependency
	primaryFile := getPrimaryFile(version.Files)
	if primaryFile == nil {
		log.Warn(fmt.Sprintf("No primary file found for %s", project.Title))
		return nil // Continue with next dependency
	}

	dep.ProjectId = project.ID
	dep.Version = version.VersionNumber
	dep.FileName = primaryFile.Filename
	dep.FileHash = primaryFile.Hashes["sha512"]
	dep.DownloadUrl = primaryFile.URL

	// Add the resolved project ID to the manifest map to prevent duplicate downloads
	manifestProjectIds[project.ID] = true

	if err = downloadAndPlace(dep, depType, rootDir, ftpClient, cache, newCache); err != nil {
		return err
	}

	// Typewriter extensions
	if typewriterMeta, ok := dep.Metadata["plugin.typewriter"]; ok {
		extensions, ok := typewriterMeta.(map[string]interface{})["extensions"].([]interface{})
		if !ok {
			log.Warn("Invalid 'extensions' format in plugin.typewriter metadata")
			return nil
		}
		for _, ext := range extensions {
			extName, ok := ext.(string)
			if !ok {
				log.Warn("Invalid extension name in plugin.typewriter metadata")
				continue
			}
			extFile := findFileByName(version.Files, extName+".jar")
			if extFile == nil {
				log.Warn(fmt.Sprintf("Extension '%s' not found for Typewriter", extName))
				continue
			}
			extDep := &manifest.Dependency{
				ProjectId:   project.ID,
				Version:     version.VersionNumber,
				FileName:    extFile.Filename,
				FileHash:    extFile.Hashes["sha512"],
				DownloadUrl: extFile.URL,
				SaveAs:      extFile.Filename,
			}
			err := downloadAndPlace(extDep, filepath.Join(depType, "Typewriter", "extensions"), rootDir, ftpClient, cache, newCache)
			if err != nil {
				return err
			}
		}
	}

	// Other dependencies
	dep.Dependencies, err = api.GetRequiredDependencies(version, m.Server, dep.DownloadIncompatible)
	if err != nil {
		return err
	}
	if len(dep.Dependencies) > 0 {
		// Filter out dependencies that are already defined in the manifest
		filteredDeps := make([]*manifest.Dependency, 0)
		for _, subDep := range dep.Dependencies {
			if subDep.ProjectId == "" {
				// If project ID is not set, include the dependency
				filteredDeps = append(filteredDeps, subDep)
			} else if !manifestProjectIds[subDep.ProjectId] {
				// Only include if not already in manifest
				filteredDeps = append(filteredDeps, subDep)
			} else {
				log.Debug(fmt.Sprintf("Skipping dependency %s (project ID: %s) as it's already defined in manifest", subDep.SaveAs, subDep.ProjectId))
			}
		}

		if len(filteredDeps) > 0 {
			log.Task(fmt.Sprintf("Processing required dependencies for %s", project.Title))
			err := processDependencies(filteredDeps, depType, m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func processHangarDependency(dep *manifest.Dependency, meta any, depType string, m *manifest.Manifest, rootDir string, ftpClient *ftp.ServerConn, cache, newCache map[string]string, manifestProjectIds map[string]bool) error {
	hangarMeta, ok := meta.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid hangar metadata format for %s", dep.SaveAs)
	}

	var projectSlug string
	var projectSlugFound bool
	for key, value := range hangarMeta {
		if strings.EqualFold(key, "projectId") || strings.EqualFold(key, "projectSlug") {
			projectSlug, projectSlugFound = value.(string)
			break
		}
	}

	if !projectSlugFound {
		return fmt.Errorf("projectSlug not found or not a string in hangar manifest metadata for %s", dep.SaveAs)
	}

	log.Task(fmt.Sprintf("Processing %s: %s", depType, projectSlug))

	project, err := api.GetHangarProject(projectSlug)
	if err != nil {
		return err
	}

	var versions []api.HangarVersion
	if dep.DownloadIncompatible {
		versions, err = api.GetAllHangarVersionsFor(project)
	} else {
		versions, err = api.GetHangarVersionsFor(project, m.Server)
	}

	if err != nil || len(versions) == 0 {
		log.Warn(fmt.Sprintf("No compatible versions found for %s", dep.SaveAs))
		return nil // Continue with next dependency
	}

	version := api.ResolveHangarVersion(versions, dep.WantedVersion)
	if version == nil {
		log.Warn(fmt.Sprintf("Wanted version '%s' not found for %s", dep.WantedVersion, dep.SaveAs))
		return nil // Continue with next dependency
	}

	// Get download URL and filename
	downloadUrl, filename, err := api.GetHangarDownloadUrl(project, version, m.Server)
	if err != nil {
		return err
	}

	dep.ProjectId = fmt.Sprintf("%d", project.ProjectID)
	dep.Version = version.Name
	dep.FileName = filename
	dep.FileHash = "" // Hangar doesn't provide hashes in the API response
	dep.DownloadUrl = downloadUrl

	// Add the resolved project ID to the manifest map to prevent duplicate downloads
	manifestProjectIds[dep.ProjectId] = true

	if err = downloadAndPlace(dep, depType, rootDir, ftpClient, cache, newCache); err != nil {
		return err
	}

	// Handle dependencies (limited support for now)
	dep.Dependencies, err = api.GetHangarRequiredDependencies(project, version, m.Server, dep.DownloadIncompatible)
	if err != nil {
		return err
	}
	if len(dep.Dependencies) > 0 {
		// Filter out dependencies that are already defined in the manifest
		filteredDeps := make([]*manifest.Dependency, 0)
		for _, subDep := range dep.Dependencies {
			if subDep.ProjectId == "" {
				// If project ID is not set, include the dependency
				filteredDeps = append(filteredDeps, subDep)
			} else if !manifestProjectIds[subDep.ProjectId] {
				// Only include if not already in manifest
				filteredDeps = append(filteredDeps, subDep)
			} else {
				log.Debug(fmt.Sprintf("Skipping dependency %s (project ID: %s) as it's already defined in manifest", subDep.SaveAs, subDep.ProjectId))
			}
		}

		if len(filteredDeps) > 0 {
			log.Task(fmt.Sprintf("Processing required dependencies for %s", project.Name))
			err := processDependencies(filteredDeps, depType, m, rootDir, ftpClient, cache, newCache, manifestProjectIds)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadAndPlace(dep *manifest.Dependency, dest string, rootDir string, ftpClient *ftp.ServerConn, cache, newCache map[string]string) error {
	finalFileName := dep.CanonicalFileName()
	finalPath := filepath.Join(dest, finalFileName)
	cacheKey := fmt.Sprintf("%s:%s", dep.ProjectId, finalFileName)
	newCache[cacheKey] = finalPath

	// Check cache
	if oldFile, ok := cache[cacheKey]; ok {
		if oldFile == finalPath {
			log.Debug(fmt.Sprintf("File %s is already up to date", finalPath))
			return nil
		}
		// Remove old file
		log.Task(fmt.Sprintf("Purging old %s", filepath.Base(oldFile)))
		removeFile(oldFile, ftpClient)
	}

	// Download
	log.Task(fmt.Sprintf("Downloading %s to %s", dep.FileName, finalPath))
	tmpPath := filepath.Join(os.TempDir(), dep.FileName)
	if err := api.DownloadFile(dep.DownloadUrl, tmpPath); err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	// Place file
	if ftpClient != nil {
		file, err := os.Open(tmpPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Ensure directory exists (relative to current FTP directory)
		if dest != "." && dest != "" {
			// Create directory structure recursively
			dirs := strings.Split(filepath.ToSlash(dest), "/")
			for _, dir := range dirs {
				if dir == "" {
					continue
				}

				// Try to change to directory, create if it doesn't exist
				if err := ftpClient.ChangeDir(dir); err != nil {
					if err := ftpClient.MakeDir(dir); err != nil {
						return fmt.Errorf("failed to create FTP directory %s: %w", dir, err)
					}
					if err := ftpClient.ChangeDir(dir); err != nil {
						return fmt.Errorf("failed to change to created directory %s: %w", dir, err)
					}
				}
			}

			// Go back to the base directory (where we started)
			for range dirs {
				if err := ftpClient.ChangeDir(".."); err != nil {
					return fmt.Errorf("failed to change back to parent directory: %w", err)
				}
			}
		}

		// Upload file using relative path
		return ftpClient.Stor(filepath.ToSlash(finalPath), file)
	} else {
		destPath := filepath.Join(rootDir, finalPath)
		destDir := filepath.Dir(destPath)

		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return err
		}

		// os.Rename may fail if the destination exists, so we remove it first
		if _, err := os.Stat(destPath); err == nil {
			if err := os.Remove(destPath); err != nil {
				return fmt.Errorf("failed to remove existing file at %s: %w", destPath, err)
			}
		}

		return os.Rename(tmpPath, destPath)
	}
}

func getPrimaryFile(files []api.ModrinthFile) *api.ModrinthFile {
	for i := range files {
		if files[i].Primary {
			return &files[i]
		}
	}
	return nil
}

func findFileByName(files []api.ModrinthFile, name string) *api.ModrinthFile {
	for i := range files {
		if strings.EqualFold(files[i].Filename, name) {
			return &files[i]
		}
	}
	return nil
}

func readCache(rootDir string, ftpClient *ftp.ServerConn) map[string]string {
	cache := make(map[string]string)
	var cachePath string
	var data []byte
	var err error

	if ftpClient != nil {
		// Use relative path for FTP
		cachePath = cacheFileName
		r, err := ftpClient.Retr(cachePath)
		if err != nil {
			return cache // Not found is ok
		}
		defer r.Close()
		data, err = io.ReadAll(r)
	} else {
		cachePath = filepath.Join(rootDir, cacheFileName)
		data, err = os.ReadFile(cachePath)
	}

	if err == nil {
		json.Unmarshal(data, &cache)
	}
	return cache
}

func writeCache(rootDir string, ftpClient *ftp.ServerConn, cache map[string]string) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	if ftpClient != nil {
		// Use relative path for FTP
		reader := strings.NewReader(string(data))
		return ftpClient.Stor(cacheFileName, reader)
	} else {
		cachePath := filepath.Join(rootDir, cacheFileName)
		return os.WriteFile(cachePath, data, 0o644)
	}
}

func removeFile(path string, ftpClient *ftp.ServerConn) {
	var err error
	if ftpClient != nil {
		ftpPath := filepath.ToSlash(path)

		if currentDir, pwdErr := ftpClient.CurrentDir(); pwdErr == nil {
			log.Debug(fmt.Sprintf("FTP current directory: %s, attempting to delete: %s", currentDir, ftpPath))
		}

		err = ftpClient.Delete(ftpPath)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to remove old file %s: %s", ftpPath, err))
		} else {
			log.Debug(fmt.Sprintf("Successfully removed %s", filepath.Base(ftpPath)))
		}
	} else {
		err = os.Remove(path)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to remove old file %s: %s", path, err))
		} else {
			log.Debug(fmt.Sprintf("Successfully removed %s", filepath.Base(path)))
		}
	}
}
