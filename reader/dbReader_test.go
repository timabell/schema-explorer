package reader

import (
	"reflect"
	"testing"
)

func TestDbValueToString(t *testing.T) {
	type args struct {
		colData  interface{}
		dataType string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "numeric with 4 bytes", want: "Type Error! []uint8",
			args: args{colData: []uint8{48, 48, 48, 48}, dataType: "numeric"},
		},
		{
			name: "numeric with float", want: "10.5",
			args: args{colData: 10.5, dataType: "numeric"},
		},
		{
			name: "numeric with float64", want: "10.5123115235",
			args: args{colData: float64(10.5123115235), dataType: "numeric"},
		},
		{
			name: "numeric with 4 bytes", want: "Type Error! []uint8",
			args: args{colData: []uint8{48, 48, 48, 48}, dataType: "numeric"},
		},
		{
			name: "ntext with string", want: "string value",
			args: args{colData: "string value", dataType: "ntext"},
		},
		{
			name: "ntext with bytes", want: "another value",
			args: args{colData: []byte("another value"), dataType: "ntext"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DbValueToString(tt.args.colData, tt.args.dataType); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("DbValueToString() = %v, want %v", *got, tt.want)
			}
		})
	}
}
