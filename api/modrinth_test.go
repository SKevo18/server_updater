package api_test

import (
	"fmt"
	"testing"

	"github.com/SKevo18/server_updater/api"
	"github.com/SKevo18/server_updater/manifest"
)

func TestFindVersionFor(t *testing.T) {
	dependency, err := api.FindModrinthVersionFor("viaversion", "@latest", manifest.Server{
		Loader:           "paper",
		MinecraftVersion: "1.21.7",
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(dependency)
}
