package aws

import (
	"context"

	"github.com/puppetlabs/wash/plugin"
)

// ec2InstanceConsoleOutput represents an EC2 instance's console output
type ec2InstanceConsoleOutput struct {
	plugin.EntryBase
	inst    *ec2Instance
	latest  bool
	content []byte
}

func newEC2InstanceConsoleOutput(ctx context.Context, inst *ec2Instance, latest bool) (*ec2InstanceConsoleOutput, error) {
	cl := &ec2InstanceConsoleOutput{}
	cl.inst = inst
	cl.latest = latest

	if cl.latest {
		cl.EntryBase = plugin.NewEntry("console-latest.out")
	} else {
		cl.EntryBase = plugin.NewEntry("console.out")
	}

	output, err := cl.inst.cachedConsoleOutput(ctx, cl.latest)
	if err != nil {
		return nil, err
	}

	cl.
		Attributes().
		SetCrtime(output.mtime).
		SetMtime(output.mtime).
		SetCtime(output.mtime).
		SetAtime(output.mtime).
		SetSize(uint64(len(output.content)))

	return cl, nil
}

func (cl *ec2InstanceConsoleOutput) Schema() *plugin.EntrySchema {
	return plugin.NewEntrySchema(cl, "console.out")
}

func (cl *ec2InstanceConsoleOutput) Read(ctx context.Context, p []byte, off int64) (int, error) {
	output, err := cl.inst.cachedConsoleOutput(ctx, cl.latest)
	if err != nil {
		return 0, err
	}

	cl.Attributes().SetSize(uint64(len(output.content)))
	return copy(p, output.content[off:]), nil
}
