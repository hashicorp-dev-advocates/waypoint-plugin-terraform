package platform

import (
	"bytes"
	"context"
	"io/ioutil"
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
	backend "remote" {
		hostname = "app.terraform.io"
		organization = "{{ .Organization }}"
		workspaces {
			name = "{{ .Workspace }}"
		}
	}
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
	GenerateConfig(*DeployConfig) error
	Init() error
	Apply() (*tfjson.State, error)
}

type TerraformImpl struct {
	version    string
	workingDir string
	terraform  *tfexec.Terraform
}

func NewTerraform(version string, workingDir string) (Terraform, error) {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: v.Must(v.NewVersion(version)),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return nil, err
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

func (t *TerraformImpl) GenerateConfig(config *DeployConfig) error {
	tmpl, err := template.New("main").Parse(terraformTemplate)
	if err != nil {
		return err
	}

	var data bytes.Buffer
	if err := tmpl.Execute(&data, config); err != nil {
		return err
	}

	err = ioutil.WriteFile("main.tf", data.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
