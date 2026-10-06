package promotetypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SigNoz/signoz/pkg/types/telemetrytypes"
)

func TestNewMaterializedPromotePath(t *testing.T) {
	testCases := []struct {
		name   string
		key    *telemetrytypes.TelemetryFieldKey
		wantOK bool
	}{
		{
			name:   "StringKey_DefaultIndex",
			key:    &telemetrytypes.TelemetryFieldKey{Name: "user.id", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
			wantOK: true,
		},
		{
			name: "NumberKey_Skipped",
			key:  &telemetrytypes.TelemetryFieldKey{Name: "retry.count", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeNumber},
		},
		{
			name: "DefaultMaterializedKey_Skipped",
			key:  &telemetrytypes.TelemetryFieldKey{Name: "http.route", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
		},
		{
			name: "CardinalKey_Skipped",
			key:  &telemetrytypes.TelemetryFieldKey{Name: "session.550e8400-e29b-41d4-a716-446655440000", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path, ok := NewMaterializedPromotePath(NewTracesAttributesTarget(), testCase.key)
			require.Equal(t, testCase.wantOK, ok)
			if !ok {
				return
			}
			assert.Equal(t, &PromotePath{
				Signal:  "traces",
				Context: "attribute",
				Path:    testCase.key.Name,
				Indexes: []WrappedIndex{{FieldDataType: telemetrytypes.FieldDataTypeString, Type: DefaultMaterializedIndexType, Granularity: DefaultMaterializedIndexGranularity, JSONDataType: telemetrytypes.String}},
			}, path)
		})
	}
}

func TestDropExistingIndexes(t *testing.T) {
	testCases := []struct {
		name     string
		existing []telemetrytypes.TelemetryFieldKeySkipIndex
		want     []WrappedIndex
	}{
		{
			name:     "SameTypeOtherParams_Dropped",
			existing: []telemetrytypes.TelemetryFieldKeySkipIndex{{IndexType: "bloom_filter(0.05)"}},
			want:     []WrappedIndex{{Type: "tokenbf_v1(1024, 2, 0)", Granularity: 1}},
		},
		{
			name:     "OtherType_Kept",
			existing: []telemetrytypes.TelemetryFieldKeySkipIndex{{IndexType: "ngrambf_v1(4, 1024, 2, 0)"}},
			want:     []WrappedIndex{{Type: "bloom_filter(0.01)", Granularity: 64}, {Type: "tokenbf_v1(1024, 2, 0)", Granularity: 1}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path := &PromotePath{Indexes: []WrappedIndex{{Type: "bloom_filter(0.01)", Granularity: 64}, {Type: "tokenbf_v1(1024, 2, 0)", Granularity: 1}}}
			path.DropExistingIndexes(testCase.existing)
			assert.Equal(t, testCase.want, path.Indexes)
		})
	}
}
