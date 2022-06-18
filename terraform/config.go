package terraform

type Config struct {
	Version string  `hcl:"version,optional"`
	Backend Backend `hcl:"backend,block"`

	Modules []Module `hcl:"module,block"`
}

type Backend struct {
	Type string `hcl:"type"`

	// type = cloud
	Organization string `hcl:"organization,optional"`
	Workspace    string `hcl:"workspace,optional"`

	// type = consul
	Address string `hcl:"address,optional"`
	Scheme  string `hcl:"scheme,optional"`
	Path    string `hcl:"path,optional"`
}

type Module struct {
	Name    string `hcl:"name"`
	Source  string `hcl:"source"`
	Version string `hcl:"version"`
	Inputs  string `hcl:"inputs,optional"`
}
