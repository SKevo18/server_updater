package manifest

type FTP struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	RemotePath string `json:"remotePath"`
}

type Manifest struct {
	FTP     *FTP     `json:"ftp"`
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
