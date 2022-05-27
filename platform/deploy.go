package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	sdk "github.com/hashicorp/waypoint-plugin-sdk/proto/gen"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/kr/pretty"
)

type DeployConfig struct {
	Organization string   `hcl:"organization"`
	Workspace    string   `hcl:"workspace"`
	Version      string   `hcl:"version"`
	Modules      []Module `hcl:"module,block"`
}

type Module struct {
	Name    string `hcl:"name"`
	Source  string `hcl:"source"`
	Version string `hcl:"version"`
	Inputs  string `hcl:"inputs,optional"`
}

type Platform struct {
	config DeployConfig
}

// Implement Configurable
func (p *Platform) Config() (interface{}, error) {
	return &p.config, nil
}

// Implement ConfigurableNotify
func (p *Platform) ConfigSet(config interface{}) error {
	_, ok := config.(*DeployConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *DeployConfig as parameter")
	}

	return nil
}

// Implement Builder
func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) StatusFunc() interface{} {
	return p.status
}

func (d *Platform) deploy(
	ctx context.Context,
	ui terminal.UI,
	log hclog.Logger,
	ji *component.JobInfo,
	dcr *component.DeclaredResourcesResp,
) (*Output, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Deploy application")

	u.Update(pretty.Sprint(ji))

	u.Update("Installing terraform")
	terraform, err := NewTerraform(d.config.Version, "./")
	if err != nil {
		return nil, err
	}

	u.Update("Generating configuration")
	err = terraform.GenerateConfig(&d.config)
	if err != nil {
		return nil, err
	}

	u.Update("Initializing workspace")
	err = terraform.Init()
	if err != nil {
		return nil, err
	}

	u.Update("Applying configuration")
	state, err := terraform.Apply()
	if err != nil {
		return nil, err
	}

	u.Update(pretty.Sprint(state))

	u.Update("Application deployed")

	result := &Output{
		Organization: d.config.Organization,
		Workspace:    d.config.Workspace,
	}
	return result, nil
}

func (d *Platform) status(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
	output *Output,
) (*sdk.StatusReport, error) {
	sg := ui.StepGroup()
	s := sg.Add("Checking the status of the deployment...")

	s.Update("Deployment is currently not implemented!")
	s.Done()

	return &sdk.StatusReport{}, nil
}
