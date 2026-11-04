package gpu

import "fmt"

type PipelineStage struct {
	Rank        int
	DeviceIndex int
	LayerStart  int
	LayerEnd    int
}

type PipelinePlan struct {
	Stages      []PipelineStage
	MicroBatch  int
	TotalLayers int
}

// BuildPipeline partitions totalLayers across nStages and assigns devices from topology.
func BuildPipeline(top *Topology, totalLayers, nStages, microBatch int) (*PipelinePlan, error) {
	if top == nil || len(top.Devices) == 0 {
		return nil, fmt.Errorf("topology contains no available compute devices")
	}
	if totalLayers < 1 {
		totalLayers = 1
	}
	if nStages < 1 {
		nStages = 1
	}
	if microBatch < 1 {
		microBatch = 4
	}

	// In single-device mode, enforce 1 stage
	if top.SingleDevice || len(top.Devices) == 1 {
		nStages = 1
	}

	stages := make([]PipelineStage, nStages)
	layersPerStage := totalLayers / nStages
	remainder := totalLayers % nStages

	currentLayer := 0
	for i := 0; i < nStages; i++ {
		extra := 0
		if i < remainder {
			extra = 1
		}
		span := layersPerStage + extra
		if span < 1 && totalLayers > 0 {
			span = 1
		}

		devIdx := i % len(top.Devices)
		stages[i] = PipelineStage{
			Rank:        i,
			DeviceIndex: top.Devices[devIdx].Index,
			LayerStart:  currentLayer,
			LayerEnd:    currentLayer + span,
		}
		currentLayer += span
	}

	return &PipelinePlan{
		Stages:      stages,
		MicroBatch:  microBatch,
		TotalLayers: totalLayers,
	}, nil
}
