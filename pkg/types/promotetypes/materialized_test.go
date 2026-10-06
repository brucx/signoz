package promotetypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SigNoz/signoz/pkg/types/telemetrytypes"
)

func TestNewMaterializedPromotePath(t *testing.T) {
	defaultIndex := WrappedIndex{FieldDataType: telemetrytypes.FieldDataTypeString, Type: DefaultMaterializedIndexType, Granularity: DefaultMaterializedIndexGranularity, JSONDataType: telemetrytypes.String}

	testCases := []struct {
		name        string
		target      Target
		key         *telemetrytypes.TelemetryFieldKey
		wantOK      bool
		wantIndexes []WrappedIndex
	}{
		{
			name:   "ColumnIndexes_Copied",
			target: NewTracesAttributesTarget(),
			key: &telemetrytypes.TelemetryFieldKey{Name: "user.id", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString, Indexes: []telemetrytypes.TelemetryFieldKeySkipIndex{
				{IndexType: "bloom_filter(0.01)", Granularity: 4},
				{IndexType: "tokenbf_v1(1024, 2, 0)", Granularity: 1},
			}},
			wantOK: true,
			wantIndexes: []WrappedIndex{
				{FieldDataType: telemetrytypes.FieldDataTypeString, Type: "bloom_filter(0.01)", Granularity: 4, JSONDataType: telemetrytypes.String},
				{FieldDataType: telemetrytypes.FieldDataTypeString, Type: "tokenbf_v1(1024, 2, 0)", Granularity: 1, JSONDataType: telemetrytypes.String},
			},
		},
		{
			name:        "NoColumnIndex_DefaultIndex",
			target:      NewTracesAttributesTarget(),
			key:         &telemetrytypes.TelemetryFieldKey{Name: "user.id", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
			wantOK:      true,
			wantIndexes: []WrappedIndex{defaultIndex},
		},
		{
			name:   "UnsupportedColumnIndex_DefaultIndex",
			target: NewTracesAttributesTarget(),
			key: &telemetrytypes.TelemetryFieldKey{Name: "user.id", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString, Indexes: []telemetrytypes.TelemetryFieldKeySkipIndex{
				{IndexType: "set(100)", Granularity: 4},
			}},
			wantOK:      true,
			wantIndexes: []WrappedIndex{defaultIndex},
		},
		{
			name:        "ResourceKey_Indexed",
			target:      NewLogsResourceTarget(),
			key:         &telemetrytypes.TelemetryFieldKey{Name: "k8s.pod.name", FieldContext: telemetrytypes.FieldContextResource, FieldDataType: telemetrytypes.FieldDataTypeString},
			wantOK:      true,
			wantIndexes: []WrappedIndex{defaultIndex},
		},
		{
			name:   "NumberKey_Skipped",
			target: NewTracesAttributesTarget(),
			key:    &telemetrytypes.TelemetryFieldKey{Name: "retry.count", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeNumber},
		},
		{
			name:   "DefaultAttribute_Skipped",
			target: NewTracesAttributesTarget(),
			key:    &telemetrytypes.TelemetryFieldKey{Name: "http.route", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
		},
		{
			name:   "DefaultResource_Skipped",
			target: NewTracesResourceTarget(),
			key:    &telemetrytypes.TelemetryFieldKey{Name: "service.name", FieldContext: telemetrytypes.FieldContextResource, FieldDataType: telemetrytypes.FieldDataTypeString},
		},
		{
			name:   "CardinalPath_Skipped",
			target: NewTracesAttributesTarget(),
			key:    &telemetrytypes.TelemetryFieldKey{Name: "session.550e8400-e29b-41d4-a716-446655440000", FieldContext: telemetrytypes.FieldContextAttribute, FieldDataType: telemetrytypes.FieldDataTypeString},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path, ok := NewMaterializedPromotePath(testCase.target, testCase.key)
			require.Equal(t, testCase.wantOK, ok)
			if !ok {
				return
			}
			assert.Equal(t, testCase.target.Entry.Signal.StringValue(), path.Signal)
			assert.Equal(t, testCase.target.Entry.FieldContext.StringValue(), path.Context)
			assert.Equal(t, testCase.key.Name, path.Path)
			assert.False(t, path.Promote)
			assert.Equal(t, testCase.wantIndexes, path.Indexes)
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

func TestValidateAndSetDefaultsIndexOnlyTargets(t *testing.T) {
	testCases := []struct {
		name    string
		target  Target
		path    *PromotePath
		wantErr bool
	}{
		{
			name:   "ResourceStringIndex_Accepted",
			target: NewTracesResourceTarget(),
			path:   &PromotePath{Signal: "traces", Context: "resource", Path: "k8s.pod.name", Indexes: []WrappedIndex{{FieldDataType: telemetrytypes.FieldDataTypeString, Type: "bloom_filter(0.01)", Granularity: 64}}},
		},
		{
			name:    "ResourcePromote_Rejected",
			target:  NewLogsResourceTarget(),
			path:    &PromotePath{Signal: "logs", Context: "resource", Path: "k8s.pod.name", Promote: true},
			wantErr: true,
		},
		{
			name:    "ScopeNumberIndex_Rejected",
			target:  NewTracesScopeTarget(),
			path:    &PromotePath{Signal: "traces", Context: "scope", Path: "lib.version", Indexes: []WrappedIndex{{FieldDataType: telemetrytypes.FieldDataTypeInt64, Type: "minmax", Granularity: 1}}},
			wantErr: true,
		},
		{
			name:    "ScopeColumnPrefix_Rejected",
			target:  NewTracesScopeTarget(),
			path:    &PromotePath{Signal: "traces", Context: "scope", Path: "scope.attributes.lib.version"},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.path.ValidateAndSetDefaults(testCase.target)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
