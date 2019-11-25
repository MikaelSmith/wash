package gcp

import (
	"context"

	"github.com/puppetlabs/wash/activity"
	"github.com/puppetlabs/wash/plugin"
	compute "google.golang.org/api/compute/v1"
)

type computeInstanceConsoleOutput struct {
	plugin.EntryBase
	instance *compute.Instance
	service  computeProjectService
	console  string
}

func newComputeInstanceConsoleOutput(inst *compute.Instance, c computeProjectService) *computeInstanceConsoleOutput {
	return &computeInstanceConsoleOutput{
		EntryBase: plugin.NewEntry("console.out"),
		instance:  inst,
		service:   c,
	}
}

func (cl *computeInstanceConsoleOutput) Schema() *plugin.EntrySchema {
	return plugin.NewEntrySchema(cl, "console.out").IsSingleton()
}

func (cl *computeInstanceConsoleOutput) Read(ctx context.Context, p []byte, off int64) (int, error) {
	if cl.console == "" {
		zone := getZone(cl.instance)
		activity.Record(ctx,
			"Getting output for instance %v in project %v, zone %v", cl.instance.Name, cl.service.projectID, zone)
		outputCall := cl.service.Instances.GetSerialPortOutput(cl.service.projectID, zone, cl.instance.Name)
		outputResp, err := outputCall.Context(ctx).Do()
		if err != nil {
			return 0, err
		}
		cl.console = outputResp.Contents
	}

	return copy(p, cl.console[off:]), nil
}
