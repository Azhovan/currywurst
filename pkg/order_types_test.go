package pkg

import "testing"

func TestGetOrderType(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantType OrderType
	}{
		{
			name:     "valid order type",
			typeName: "vegan",
			wantType: Vegan{},
		},
		{
			name:     "invalid order type",
			typeName: "blah",
			wantType: nil,
		},
		{
			name:     "empty order type name",
			typeName: "",
			wantType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType := GetOrderType(tt.typeName)
			if gotType != tt.wantType {
				t.Errorf("GetOrderType(%q) = %v, want:%v", tt.typeName, gotType, tt.wantType)
			}
		})
	}
}
