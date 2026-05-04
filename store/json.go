package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JamesPagetButler/wyrd/model"
)

// JSONFile is a Crawl-phase persistence layer that round-trips a Graph
// through a single JSON file.
//
// Format (top-level object):
//
//	{
//	  "version":    1,
//	  "nodes":      [<Node>...],
//	  "hyperedges": [<Hyperedge>...]
//	}
//
// Loading rebuilds the in-memory incidence index from the hyperedges.
type JSONFile struct {
	Path string
}

// jsonGraph is the on-disk envelope.
type jsonGraph struct {
	Version    int                `json:"version"`
	Nodes      []model.Node       `json:"nodes"`
	Hyperedges []model.Hyperedge  `json:"hyperedges"`
}

const currentVersion = 1

// Save serialises the graph to the configured Path.
func (j JSONFile) Save(g *model.Graph) error {
	if j.Path == "" {
		return fmt.Errorf("wyrd: store: empty path")
	}
	envelope := jsonGraph{
		Version:    currentVersion,
		Nodes:      g.Nodes(),
		Hyperedges: g.Hyperedges(),
	}
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("wyrd: store: marshal: %w", err)
	}
	if err := os.WriteFile(j.Path, data, 0o644); err != nil {
		return fmt.Errorf("wyrd: store: write %s: %w", j.Path, err)
	}
	return nil
}

// Load reads the graph from the configured Path. Returns a fresh
// *model.Graph with all nodes and hyperedges restored and the
// incidence index rebuilt.
func (j JSONFile) Load() (*model.Graph, error) {
	if j.Path == "" {
		return nil, fmt.Errorf("wyrd: store: empty path")
	}
	data, err := os.ReadFile(j.Path)
	if err != nil {
		return nil, fmt.Errorf("wyrd: store: read %s: %w", j.Path, err)
	}
	var envelope jsonGraph
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("wyrd: store: unmarshal: %w", err)
	}
	if envelope.Version != currentVersion {
		return nil, fmt.Errorf("wyrd: store: unsupported envelope version %d (want %d)",
			envelope.Version, currentVersion)
	}

	g := model.NewGraph()
	for _, n := range envelope.Nodes {
		if err := g.AddNode(n); err != nil {
			return nil, fmt.Errorf("wyrd: store: load node %s: %w", n.ID, err)
		}
	}
	for _, e := range envelope.Hyperedges {
		if err := g.AddHyperedge(e); err != nil {
			return nil, fmt.Errorf("wyrd: store: load edge %s: %w", e.ID, err)
		}
	}
	return g, nil
}
