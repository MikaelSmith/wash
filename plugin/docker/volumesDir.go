package docker

import (
	"context"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/puppetlabs/wash/activity"
	"github.com/puppetlabs/wash/plugin"
)

type volumesDir struct {
	plugin.EntryBase
	client *client.Client
}

func newVolumesDir(client *client.Client) *volumesDir {
	volumesDir := &volumesDir{
		EntryBase: plugin.NewEntry("volumes"),
	}
	volumesDir.client = client
	return volumesDir
}

func (vs *volumesDir) Schema() *plugin.EntrySchema {
	return plugin.NewEntrySchema(vs, "volumes").IsSingleton()
}

func (vs *volumesDir) ChildSchemas() []*plugin.EntrySchema {
	return []*plugin.EntrySchema{
		(&volume{}).Schema(),
	}
}

// List
func (vs *volumesDir) List(ctx context.Context) ([]plugin.Entry, error) {
	volumes, err := vs.client.VolumeList(ctx, filters.Args{})
	if err != nil {
		return nil, err
	}

	activity.Record(ctx, "Listing %v volumes in %v", len(volumes.Volumes), vs)
	keys := make([]plugin.Entry, len(volumes.Volumes))
	for i, inst := range volumes.Volumes {
		if keys[i], err = newVolume(vs.client, inst); err != nil {
			return nil, err
		}
	}
	return keys, nil
}
