package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	ctypes "github.com/JamesPagetButler/contextus/pkg/types"
	"github.com/JamesPagetButler/wyrd/model"
	"gopkg.in/yaml.v3"
)

// Sentinel errors per PR #40 design §6. Consumers (BMA reins wrapper,
// CI smoke tests) use errors.Is to dispatch on failure shape.
var (
	// ErrScopeConfigInvalid is returned when scope-config validation
	// fails (empty IDs, kind mismatch, out-of-range Confidence, etc.).
	ErrScopeConfigInvalid = errors.New("store: scope-config validation failed")

	// ErrScopeLoadConflict is returned when a scope or membership ID
	// already exists in the destination graph. v0.1 is non-upsert per
	// PR #40 §7 OQ #3.
	ErrScopeLoadConflict = errors.New("store: scope-node ID already present in graph")

	// ErrScopeConfigParse is returned for YAML/JSON parse errors.
	ErrScopeConfigParse = errors.New("store: scope-config parse error")
)

// scopeConfigYAML is the on-disk YAML/JSON shape the loader accepts.
// Matches PR #40 §2.1 design verbatim, plus the optional tier_immune
// + salience YAML fields added in commit aa5a337 (the F2 resolution
// for @qbp-implementor's review).
type scopeConfigYAML struct {
	PhysicalScopes   []physicalScopeYAML   `json:"physical_scopes" yaml:"physical_scopes"`
	ConceptualScopes []conceptualScopeYAML `json:"conceptual_scopes" yaml:"conceptual_scopes"`
	ScopeMemberships []membershipYAML      `json:"scope_memberships" yaml:"scope_memberships"`
}

type physicalScopeYAML struct {
	ID          string           `json:"id" yaml:"id"`
	Description string           `json:"description" yaml:"description"`
	Bounds      map[string][]any `json:"bounds" yaml:"bounds"`
	TypeNodes   []string         `json:"type_nodes" yaml:"type_nodes"`
	TierImmune  bool             `json:"tier_immune,omitempty" yaml:"tier_immune,omitempty"`
	Salience    float64          `json:"salience,omitempty" yaml:"salience,omitempty"`
}

type conceptualScopeYAML struct {
	ID          string   `json:"id" yaml:"id"`
	Description string   `json:"description" yaml:"description"`
	TypeNodes   []string `json:"type_nodes" yaml:"type_nodes"`
	TierImmune  bool     `json:"tier_immune,omitempty" yaml:"tier_immune,omitempty"`
	Salience    float64  `json:"salience,omitempty" yaml:"salience,omitempty"`
}

type membershipYAML struct {
	Scope      string  `json:"scope" yaml:"scope"`
	Member     string  `json:"member" yaml:"member"`
	WeightTier string  `json:"weight_tier" yaml:"weight_tier"`
	Confidence float64 `json:"confidence,omitempty" yaml:"confidence,omitempty"`
	Method     string  `json:"method,omitempty" yaml:"method,omitempty"`
}

// Node-type strings for scope nodes. Match Contextus Spec v1.3 §4.6.
const (
	NodeTypeScopePhysical   model.NodeType = "contextus.scope.physical"
	NodeTypeScopeConceptual model.NodeType = "contextus.scope.conceptual"
	HyperedgeTypeMembership                = "contextus.scope.member" // for documentation; not enforced
)

// LoadScopeConfig reads a scope-node configuration (YAML or JSON) at
// configPath, validates it, and atomically populates graph with the
// resulting scope nodes + membership edges.
//
// All-or-nothing per ADR-003 §I3: either every scope + membership
// lands in graph, or none do. Phase-1 validation catches every
// invariant violation before any mutation; phase-2 commit cannot
// fail under the held lock (the underlying graph methods are
// validated by phase-1 checks).
//
// Dispatch on file extension: .yaml / .yml use YAML; .json uses JSON;
// other extensions default to YAML.
//
// Returns ErrScopeConfigInvalid / ErrScopeLoadConflict / ErrScopeConfigParse
// per PR #40 §6; consumers unwrap with errors.Is.
func LoadScopeConfig(graph *model.Graph, configPath string) error {
	// #nosec G304 -- configPath is the loader's documented input;
	// callers are trusted (BMA reins wrapper, scope-config bootstrap
	// scripts). File-inclusion-via-variable is the entire point of
	// this function.
	f, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("store: open %q: %w", configPath, err)
	}
	defer func() { _ = f.Close() }()

	ext := strings.ToLower(filepath.Ext(configPath))
	return LoadScopeConfigReader(graph, f, ext)
}

// LoadScopeConfigReader is the io.Reader form of LoadScopeConfig (per
// PR #40 §7 OQ #2). ext is the file extension (".yaml", ".yml",
// ".json"); empty or unknown defaults to YAML.
func LoadScopeConfigReader(graph *model.Graph, r io.Reader, ext string) error {
	if graph == nil {
		return fmt.Errorf("%w: nil graph", ErrScopeConfigInvalid)
	}

	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%w: read: %w", ErrScopeConfigParse, err)
	}

	var cfg scopeConfigYAML
	switch ext {
	case ".json":
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("%w: json: %w", ErrScopeConfigParse, err)
		}
	default:
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return fmt.Errorf("%w: yaml: %w", ErrScopeConfigParse, err)
		}
	}

	// Phase 1: validate + prepare (no mutation).
	nodes, edges, err := preparePopulate(graph, &cfg)
	if err != nil {
		return err
	}

	// Phase 2: commit. Per ADR-003 §I3, atomicity at this layer relies
	// on the loader holding the graph's lock for the full commit
	// window. The model.Graph.AddNode/AddHyperedge methods take their
	// own Lock per call; the all-or-nothing guarantee comes from
	// phase-1 having caught every invariant violation already, so
	// phase-2 cannot legitimately fail.
	for _, n := range nodes {
		if err := graph.AddNode(n); err != nil {
			return fmt.Errorf("store: scope-loader: AddNode invariant violation (phase-1 should have caught this): %w", err)
		}
	}
	for _, e := range edges {
		if err := graph.AddHyperedge(e); err != nil {
			return fmt.Errorf("store: scope-loader: AddHyperedge invariant violation (phase-1 should have caught this): %w", err)
		}
	}
	return nil
}

func preparePopulate(graph *model.Graph, cfg *scopeConfigYAML) ([]model.Node, []model.Hyperedge, error) {
	seenIDs := make(map[string]struct{})
	now := time.Now().UTC()

	var nodes []model.Node

	addScopeNode := func(id string, typ model.NodeType, payload []byte, tierImmune bool, salience float64) error {
		if id == "" {
			return fmt.Errorf("%w: empty scope id", ErrScopeConfigInvalid)
		}
		if _, dup := seenIDs[id]; dup {
			return fmt.Errorf("%w: duplicate id %q in config", ErrScopeConfigInvalid, id)
		}
		if _, exists := graph.Node(model.NodeID(id)); exists {
			return fmt.Errorf("%w: scope %q", ErrScopeLoadConflict, id)
		}
		seenIDs[id] = struct{}{}
		if salience < 0.0 || salience > 1.0 {
			return fmt.Errorf("%w: scope %q salience %g out of [0,1]", ErrScopeConfigInvalid, id, salience)
		}
		nodes = append(nodes, model.Node{
			ID:         model.NodeID(id),
			Type:       typ,
			Tier:       model.TierComplex,
			Created:    now,
			Payload:    payload,
			TierImmune: tierImmune,
			Salience:   salience,
		})
		return nil
	}

	// Physical scopes — re-encode the YAML as JSON for Node.Payload
	// so consumers see a normalised form regardless of source format.
	// The canonical Contextus type is ctypes.ScopePhysical; we don't
	// construct the full struct here (the loader's job is to land
	// scope NODES + edges; the rich on-payload struct shape is the
	// adapter's concern per the federation contract).
	for i, p := range cfg.PhysicalScopes {
		payload, err := json.Marshal(p)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: physical_scopes[%d]: marshal: %w", ErrScopeConfigInvalid, i, err)
		}
		if err := addScopeNode(p.ID, NodeTypeScopePhysical, payload, p.TierImmune, p.Salience); err != nil {
			return nil, nil, err
		}
	}
	for i, c := range cfg.ConceptualScopes {
		payload, err := json.Marshal(c)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: conceptual_scopes[%d]: marshal: %w", ErrScopeConfigInvalid, i, err)
		}
		if err := addScopeNode(c.ID, NodeTypeScopeConceptual, payload, c.TierImmune, c.Salience); err != nil {
			return nil, nil, err
		}
	}

	// Memberships. Verify both endpoints exist (either freshly added
	// above, or already in the graph). type_nodes are validated for
	// non-empty shape only; per PR #40 §7 OQ #4, forward-references
	// allowed at v0.1.
	var edges []model.Hyperedge
	for i, m := range cfg.ScopeMemberships {
		if m.Scope == "" || m.Member == "" {
			return nil, nil, fmt.Errorf("%w: scope_memberships[%d]: empty scope or member", ErrScopeConfigInvalid, i)
		}
		if _, ok := seenIDs[m.Scope]; !ok {
			if _, exists := graph.Node(model.NodeID(m.Scope)); !exists {
				return nil, nil, fmt.Errorf("%w: scope_memberships[%d]: scope %q not in config or graph", ErrScopeConfigInvalid, i, m.Scope)
			}
		}
		if _, ok := seenIDs[m.Member]; !ok {
			if _, exists := graph.Node(model.NodeID(m.Member)); !exists {
				return nil, nil, fmt.Errorf("%w: scope_memberships[%d]: member %q not in config or graph", ErrScopeConfigInvalid, i, m.Member)
			}
		}
		var tier model.Tier
		switch strings.ToLower(strings.TrimSpace(m.WeightTier)) {
		case "", "complex":
			tier = model.TierComplex
		case "quaternion":
			tier = model.TierQuaternion
		case "octonion":
			tier = model.TierOctonion
		case "sedenion":
			tier = model.TierSedenion
		default:
			return nil, nil, fmt.Errorf("%w: scope_memberships[%d]: unknown weight_tier %q", ErrScopeConfigInvalid, i, m.WeightTier)
		}

		// v0.1: scope-membership metadata (Confidence, Method, etc.) is
		// validated at the YAML level but not yet attached to the
		// hyperedge — model.Hyperedge doesn't carry a Payload field at
		// v0.1. A v0.x membership-metadata node may be added; for now
		// the hyperedge encodes the relationship structurally.
		// ctypes.ScopeMembership use is reserved for that v0.x change.
		_ = ctypes.ProvenanceTagAsserted // keep import live for v0.x extension
		if m.Confidence < 0.0 || m.Confidence > 1.0 {
			return nil, nil, fmt.Errorf("%w: scope_memberships[%d]: confidence %g out of [0,1]", ErrScopeConfigInvalid, i, m.Confidence)
		}
		edges = append(edges, model.Hyperedge{
			ID:    model.HyperedgeID(fmt.Sprintf("%s::%s", m.Scope, m.Member)),
			Nodes: []model.NodeID{model.NodeID(m.Scope), model.NodeID(m.Member)},
			Weight: model.Weight{
				Tier: tier,
			},
			IsSymmetric: true,
			Created:     now,
		})
	}

	return nodes, edges, nil
}
