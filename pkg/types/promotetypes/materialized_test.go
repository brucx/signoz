package promotetypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SigNoz/signoz/pkg/types/telemetrytypes"
)

func TestNewMaterializedPromotePath(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name: "BarePath_DefaultIndex",
			path: "user.id",
		},
		{
			name:    "CardinalPath_Rejected",
			path:    "session.550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path, err := NewMaterializedPromotePath(NewTracesAttributesTarget(), testCase.path)
			if testCase.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, &PromotePath{
				Signal:  "traces",
				Context: "attribute",
				Path:    testCase.path,
				Promote: false,
				Indexes: []WrappedIndex{{FieldDataType: telemetrytypes.FieldDataTypeString, Type: "bloom_filter(0.01)", Granularity: 64, JSONDataType: telemetrytypes.String}},
			}, path)
		})
	}
}

func TestIsDefaultMaterialized(t *testing.T) {
	testCases := []struct {
		name   string
		target Target
		path   string
		want   bool
	}{
		{name: "TracesDefaultAttribute_Default", target: NewTracesAttributesTarget(), path: "http.route", want: true},
		{name: "TracesGenAIAttribute_Default", target: NewTracesAttributesTarget(), path: "gen_ai.request.model", want: true},
		{name: "TracesCustomAttribute_NotDefault", target: NewTracesAttributesTarget(), path: "user.id"},
		{name: "LogsBodySamePath_NotDefault", target: NewLogsBodyTarget(), path: "http.route"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, testCase.target.IsDefaultMaterialized(testCase.path))
		})
	}
}
