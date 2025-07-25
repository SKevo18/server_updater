# Minecraft Server Updater

Updating Minecraft servers is a pain – everytime a new version of Minecraft is released, you have to update your server and all your plugins (or mods) manually, hand by hand, visiting the various websites to find the latest versions, only to leave disappointed if no new version is available yet.

This tool aims to make updating your server and plugins/mods as easy as possible, so you can focus on playing your favorite game.

## Features

- [] Automatically update your server.jar
  - [ ] Supports Paper and its forks (Purpur, Folia, Leaf, etc.)
  - [ ] Download a specific version, or the latest version for a specific Minecraft version
- [x] Automatically update your plugin jars
  - [x] Supports plugins from Modrinth and Hangar
  - [ ] Supports plugins published to GitHub releases and development builds from Jenkins API (if available)
  - [x] Choose a specific or latest version of the plugin
- [x] Manifest file for server and plugin definitions
- [x] Cache file to record current versions

## Usage

See [`example_server_manifest.json`](example_server_manifest.json) for an example of a server manifest.

Download from [releases](https://github.com/SKevo18/server_updater/releases)
