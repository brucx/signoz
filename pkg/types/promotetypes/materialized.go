package promotetypes

import (
	"github.com/SigNoz/signoz/pkg/types/telemetrytypes"
	"github.com/SigNoz/signoz/pkg/valuer"
)

const (
	DefaultMaterializedIndexType        = "bloom_filter(0.01)"
	DefaultMaterializedIndexGranularity = 64
)

var (
	SkipReasonDefaultMaterialized = MaterializedSkipReason{valuer.NewString("default_materialized")}
	SkipReasonNotString           = MaterializedSkipReason{valuer.NewString("not_string")}
	SkipReasonColumnNotIndexed    = MaterializedSkipReason{valuer.NewString("column_not_indexed")}
	SkipReasonInvalidPath         = MaterializedSkipReason{valuer.NewString("invalid_path")}
	SkipReasonAlreadyIndexed      = MaterializedSkipReason{valuer.NewString("already_indexed")}
)

// defaultMaterializedPaths are keyed by signal/context.
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

type MaterializedSkipReason struct {
	valuer.String
}

type SkippedMaterializedPath struct {
	Signal  string                 `json:"signal" required:"true"`
	Context string                 `json:"context" required:"true"`
	Path    string                 `json:"path" required:"true"`
	Reason  MaterializedSkipReason `json:"reason" required:"true"`
}

type IndexMaterializedPathsResult struct {
	Indexed []*PromotePath             `json:"indexed" required:"true"`
	Skipped []*SkippedMaterializedPath `json:"skipped" required:"true"`
}

type IndexMaterializedPathsParams struct {
	Signal string `query:"signal" json:"signal" required:"true"`
	DryRun bool   `query:"dryRun" json:"dryRun"`
}

func (p *IndexMaterializedPathsParams) Target() (Target, error) {
	return NewTargetFromText(p.Signal, telemetrytypes.FieldContextAttribute.StringValue())
}

func (MaterializedSkipReason) Enum() []any {
	return []any{SkipReasonDefaultMaterialized, SkipReasonNotString, SkipReasonColumnNotIndexed, SkipReasonInvalidPath, SkipReasonAlreadyIndexed}
}

func NewMaterializedPromotePath(target Target, name string) (*PromotePath, error) {
	path := &PromotePath{
		Signal:  target.Entry.Signal.StringValue(),
		Context: target.Entry.FieldContext.StringValue(),
		Path:    name,
		Promote: false,
		Indexes: []WrappedIndex{{FieldDataType: telemetrytypes.FieldDataTypeString, Type: DefaultMaterializedIndexType, Granularity: DefaultMaterializedIndexGranularity}},
	}
	if err := path.ValidateAndSetDefaults(target); err != nil {
		return nil, err
	}
	return path, nil
}

func NewSkippedMaterializedPath(target Target, name string, reason MaterializedSkipReason) *SkippedMaterializedPath {
	return &SkippedMaterializedPath{
		Signal:  target.Entry.Signal.StringValue(),
		Context: target.Entry.FieldContext.StringValue(),
		Path:    name,
		Reason:  reason,
	}
}

func (t Target) IsDefaultMaterialized(path string) bool {
	_, ok := defaultMaterializedPaths[t.Entry.Signal.StringValue()+"/"+t.Entry.FieldContext.StringValue()][path]
	return ok
}
