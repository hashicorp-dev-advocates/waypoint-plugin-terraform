package terraform

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	v "github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

const terraformTemplate = `
terraform {
	required_providers {}
	{{- if eq .Backend.Type "cloud" }}
	backend "remote" {
		hostname = "app.terraform.io"
		organization = "{{ .Backend.Organization }}"
		workspaces {
			name = "{{ .Backend.Workspace }}"
		}
	}
	{{- else if eq .Backend.Type "consul" }}
	backend "consul" {
    address = "{{ .Backend.Address }}"
    scheme  = "{{ .Backend.Scheme }}"
    path    = "{{ .Backend.Path }}"
  }
	{{- else }}
	backend "local" {}
	{{- end }}
}

{{- $modules := .Modules }}
{{- range $element := $modules}}
module "{{ .Name }}" {
	source = "{{ .Source }}"
	version = "{{ .Version }}"

	{{ .Inputs }}
}
{{- end }}

{{- range $element := $modules}}
output "{{ .Name }}" {
	value = module.{{ .Name }}
}
{{- end }}
`

type Terraform interface {
	Init() error
	Apply() (*tfjson.State, error)
	Destroy() error
	Clean() error
}

type TerraformImpl struct {
	version    string
	workingDir string
	config     *Config
	terraform  *tfexec.Terraform
}

func NewTerraform(version string, workingDir string) (Terraform, error) {
	var err error
	execPath := "/usr/bin/terraform"
	if version != "" {
		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: v.Must(v.NewVersion(version)),
		}

		execPath, err = installer.Install(context.Background())
		if err != nil {
			return nil, err
		}
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, err
	}

	return &TerraformImpl{
		version:    version,
		workingDir: workingDir,
		terraform:  tf,
	}, nil
}

func (t *TerraformImpl) Init() error {
	err := t.terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}

	return nil
}

func (t *TerraformImpl) Apply() (*tfjson.State, error) {
	opts := []tfexec.ApplyOption{}
	err := t.terraform.Apply(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	state, err := t.terraform.Show(context.Background())
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (t *TerraformImpl) Destroy() error {
	opts := []tfexec.DestroyOption{}
	err := t.terraform.Destroy(context.Background(), opts...)
	if err != nil {
		return err
	}
	return nil
}

func (t *TerraformImpl) Clean() error {
	err := os.RemoveAll(t.workingDir)
	if err != nil {
		return err
	}
	return nil
}

func GenerateConfig(config *Config) (string, error) {
	tmpl, err := template.New("main").Parse(terraformTemplate)
	if err != nil {
		return "", err
	}

	var data bytes.Buffer
	if err := tmpl.Execute(&data, config); err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp("", "waypoint-terraform")
	if err != nil {
		return "", err
	}

	file := filepath.Join(dir, "main.tf")
	err = ioutil.WriteFile(file, data.Bytes(), 0644)
	if err != nil {
		return "", err
	}

	return dir, nil
}
