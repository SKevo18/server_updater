package manifest

type Server struct {
	Loader           string   `json:"loader"`
	LoaderVersion    string   `json:"loaderVersion"`
	LoaderFile       string   `json:"loaderFile"`
	MinecraftVersion string   `json:"minecraftVersion"`
	Supports         []string `json:"supports"`
}
