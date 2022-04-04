package tunnel

type Identity struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Tunnel struct {
	Local    string `yaml:"local"`
	Remote   string `yaml:"remote"`
	Target   string `yaml:"target"`
	Identity string `yaml:"identity"`
}

type Config struct {
	Identity []*Identity `yaml:"identies"`
	Tunnels  []*Tunnel   `yaml:"tunnels"`
}
