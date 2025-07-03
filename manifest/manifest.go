package manifest

type Manifest struct {
	Server  Server   `json:"server"`
	Plugins []Plugin `json:"plugins"`
	Mods    []Mod    `json:"mods"`
}

func (m *Manifest) HasPlugins() bool {
	return len(m.Plugins) > 0
}

func (m *Manifest) HasMods() bool {
	return len(m.Mods) > 0
}
