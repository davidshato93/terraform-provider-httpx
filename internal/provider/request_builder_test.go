package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertTerraformMap(t *testing.T) {
	tests := []struct {
		name   string
		tfMap  types.Map
		want   map[string]string
		wantErr bool
	}{
		{
			name:   "null map",
			tfMap:  types.MapNull(types.StringType),
			want:   nil,
			wantErr: false,
		},
		{
			name:   "unknown map",
			tfMap:  types.MapUnknown(types.StringType),
			want:   nil,
			wantErr: false,
		},
		{
			name: "valid map",
			tfMap: types.MapValueMust(types.StringType, map[string]attr.Value{
				"key1": types.StringValue("value1"),
				"key2": types.StringValue("value2"),
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: false,
		},
		{
			name:   "empty map",
			tfMap:  types.MapValueMust(types.StringType, map[string]attr.Value{}),
			want:   map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertTerraformMap(context.Background(), tt.tfMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTerraformMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("ConvertTerraformMap() = %v, want nil", got)
				return
			}
			if tt.want != nil {
				if len(got) != len(tt.want) {
					t.Errorf("ConvertTerraformMap() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("ConvertTerraformMap() [%s] = %v, want %v", k, got[k], v)
					}
				}
			}
		})
	}
}

func TestConvertTerraformList(t *testing.T) {
	tests := []struct {
		name     string
		tfList   types.List
		converter func(interface{}) (int64, error)
		want     []int64
		wantErr  bool
	}{
		{
			name:   "null list",
			tfList: types.ListNull(types.Int64Type),
			converter: func(v interface{}) (int64, error) {
				if intVal, ok := v.(types.Int64); ok {
					return intVal.ValueInt64(), nil
				}
				return 0, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "unknown list",
			tfList: types.ListUnknown(types.Int64Type),
			converter: func(v interface{}) (int64, error) {
				if intVal, ok := v.(types.Int64); ok {
					return intVal.ValueInt64(), nil
				}
				return 0, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "valid list",
			tfList: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(2),
				types.Int64Value(3),
			}),
			converter: func(v interface{}) (int64, error) {
				if intVal, ok := v.(types.Int64); ok {
					return intVal.ValueInt64(), nil
				}
				return 0, nil
			},
			want:    []int64{1, 2, 3},
			wantErr: false,
		},
		{
			name:   "empty list",
			tfList: types.ListValueMust(types.Int64Type, []attr.Value{}),
			converter: func(v interface{}) (int64, error) {
				if intVal, ok := v.(types.Int64); ok {
					return intVal.ValueInt64(), nil
				}
				return 0, nil
			},
			want:    []int64{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertTerraformList(context.Background(), tt.tfList, tt.converter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertTerraformList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("ConvertTerraformList() = %v, want nil", got)
				return
			}
			if tt.want != nil {
				if len(got) != len(tt.want) {
					t.Errorf("ConvertTerraformList() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for i, v := range tt.want {
					if got[i] != v {
						t.Errorf("ConvertTerraformList() [%d] = %v, want %v", i, got[i], v)
					}
				}
			}
		})
	}
}

