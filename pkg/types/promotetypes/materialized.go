package promotetypes

import (
	"github.com/SigNoz/signoz/pkg/types/telemetrytypes"
)

// defaultMaterializedPaths are materialized on every install by the schema migrator, keyed by signal/context.
var defaultMaterializedPaths = map[string]map[string]struct{}{
	"traces/attribute": {
		"db.system":           {},
		"http.route":          {},
		"messaging.operation": {},
		"messaging.system":    {},
		"peer.service":        {},
		"rpc.method":          {},
		"rpc.service":         {},
		"rpc.system":          {},

		"gen_ai.agent.name":                                 {},
		"gen_ai.provider.name":                              {},
		"gen_ai.request.model":                              {},
		"gen_ai.tool.name":                                  {},
		"gen_ai.usage.cache_creation.input_tokens":          {},
		"gen_ai.usage.cache_read.input_tokens":              {},
		"gen_ai.usage.input_tokens":                         {},
		"gen_ai.usage.output_tokens":                        {},
		"signoz.gen_ai.usage.cache_read.input_tokens.cost":  {},
		"signoz.gen_ai.usage.cache_write.input_tokens.cost": {},
		"signoz.gen_ai.usage.input_tokens.cost":             {},
		"signoz.gen_ai.usage.output_tokens.cost":            {},
		"signoz.gen_ai.usage.tokens.cost":                   {},
	},
}

type IndexMaterializedPathsParams struct {
	Signal  string `query:"signal" json:"signal"`
	Context string `query:"context" json:"context"`
	DryRun  bool   `query:"dryRun" json:"dryRun"`
}

func (p *IndexMaterializedPathsParams) Filters() ListPromotedPathsFilters {
	return ListPromotedPathsFilters{Signal: p.Signal, Context: p.Context}
}

// NewMaterializedPromotePath maps a materialized string key onto an unpromoted path carrying index.
// ok is false for non-string keys and paths the target rejects.
func NewMaterializedPromotePath(target Target, key *telemetrytypes.TelemetryFieldKey, index WrappedIndex) (*PromotePath, bool) {
	if key.FieldDataType != telemetrytypes.FieldDataTypeString {
		return nil, false
	}

	path := &PromotePath{
		Signal:  target.Entry.Signal.StringValue(),
		Context: target.Entry.FieldContext.StringValue(),
		Path:    key.Name,
		Promote: false,
		Indexes: []WrappedIndex{index},
	}
	if err := path.ValidateAndSetDefaults(target); err != nil {
		return nil, false
	}
	return path, true
}

// SkipExistingIndexes removes from the indexes to create those whose type is already built on the path.
func (i *PromotePath) SkipExistingIndexes(existing []telemetrytypes.TelemetryFieldKeySkipIndex) {
	built := map[string]struct{}{}
	for _, index := range existing {
		if indexType, ok := (WrappedIndex{Type: index.IndexType}).SkipIndexType(); ok {
			built[string(indexType)] = struct{}{}
		}
	}

	indexes := make([]WrappedIndex, 0, len(i.Indexes))
	for _, index := range i.Indexes {
		indexType, _ := index.SkipIndexType()
		if _, ok := built[string(indexType)]; !ok {
			indexes = append(indexes, index)
		}
	}
	i.Indexes = indexes
}

// IsDefaultMaterialized reports whether path is materialized on every install by the schema migrator.
func (t Target) IsDefaultMaterialized(path string) bool {
	_, ok := defaultMaterializedPaths[t.Entry.Signal.StringValue()+"/"+t.Entry.FieldContext.StringValue()][path]
	return ok
}
