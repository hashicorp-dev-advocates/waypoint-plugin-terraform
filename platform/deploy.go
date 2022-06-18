package platform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eveldcorp/waypoint-plugin-terraform/terraform"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type Platform struct {
	config terraform.Config
}

// Implement Configurable
func (p *Platform) Config() (interface{}, error) {
	return &p.config, nil
}

// Implement ConfigurableNotify
func (p *Platform) ConfigSet(config interface{}) error {
	_, ok := config.(*terraform.Config)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("expected *terraform.Config as parameter")
	}

	return nil
}

// Implement Builder
func (p *Platform) DeployFunc() interface{} {
	return p.deploy
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
	u.Update("Generating configuration")
	dir, err := terraform.GenerateConfig(&d.config)
	if err != nil {
		log.Error("error generating config: %s", err)
		return nil, err
	}

	u.Update("Installing terraform")
	tf, err := terraform.NewTerraform(d.config.Version, dir)
	if err != nil {
		log.Error("error installing Terraform: %s", err)
		return nil, err
	}

	u.Update("Initializing workspace")
	err = tf.Init()
	if err != nil {
		log.Error("error initializing workspace: %s", err)
		return nil, err
	}

	u.Update("Applying configuration")
	state, err := tf.Apply()
	if err != nil {
		log.Error("error applying configuration: %s", err)
		return nil, err
	}

	u.Update("Terraform apply successful")
	tf.Clean()

	bytes, err := json.Marshal(state)
	if err != nil {
		log.Error("error converting state to bytes: %s", err)
		return nil, err
	}
	result := &Output{
		State: bytes,
	}

	return result, nil
}
