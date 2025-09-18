package currency

import "testing"

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.234567890, 1.234568},
		{0.123456789, 0.123457},
		{100.0, 100.0},
		{0.0, 0.0},
		{1.9999995, 2.0},
	}

	for _, test := range tests {
		result := FormatCurrency(test.input)
		if result != test.expected {
			t.Errorf("FormatCurrency(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}
