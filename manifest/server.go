package manifest

type Server struct {
	Loader           string   `json:"loader"`
	LoaderVersion    string   `json:"loaderVersion"`
	MinecraftVersion string   `json:"minecraftVersion"`
	Supports         []string `json:"supports"`
}
