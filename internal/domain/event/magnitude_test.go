package event_test

import (
	"testing"

	"github.com/jwgal/geopulse/internal/domain/event"
	"github.com/stretchr/testify/assert"
)

func TestNewMagnitude(t *testing.T) {
	tests := []struct {
		name      string  // description of this test case
		value     float64 // Named input parameters for target function.
		scale     string
		wantValue float64
		wantScale string
		wantErr   bool
	}{
		{
			name:      "Valid Magnitude",
			value:     5.5,
			scale:     event.MagnitudeScaleMw,
			wantValue: 5.5,
			wantScale: event.MagnitudeScaleMw,
			wantErr:   false,
		},
		{
			name:    "Invalid Magnitude Value",
			value:   -2.0,
			scale:   event.MagnitudeScaleMw,
			wantErr: true,
		},
		{
			name:    "Magnitude Value Too High",
			value:   11.0,
			scale:   event.MagnitudeScaleMw,
			wantErr: true,
		},
		{
			name:    "Magnitude Value Too Low",
			value:   -3.5,
			scale:   event.MagnitudeScaleMw,
			wantErr: true,
		},
		{
			name:      "Unknown Magnitude Scale",
			value:     5.0,
			scale:     "invalid_scale",
			wantValue: 5.0,
			wantScale: event.MagnitudeScaleUnknown,
			wantErr:   false,
		},
		{
			name:      "Maximum Valid Magnitude",
			value:     10.0,
			scale:     event.MagnitudeScaleMw,
			wantValue: 10.0,
			wantScale: event.MagnitudeScaleMw,
			wantErr:   false,
		},
		{
			name:      "case Insensitive Scale",
			value:     4.0,
			scale:     "Mw",
			wantValue: 4.0,
			wantScale: event.MagnitudeScaleMw,
			wantErr:   false,
		},
		{
			name:      "Scale with Leading/Trailing Spaces",
			value:     3.5,
			scale:     "  ml  ",
			wantValue: 3.5,
			wantScale: event.MagnitudeScaleMl,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := event.NewMagnitude(tt.value, tt.scale)
			if gotErr != nil {
				if !tt.wantErr {
					assert.Failf(t, "NewMagnitude() returned unexpected error", "error: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewMagnitude() succeeded unexpectedly")
			}
			if got.Value() != tt.wantValue || got.Scale() != tt.wantScale {
				assert.Failf(t, "NewMagnitude() mismatch", "got: %s (value: %v, scale: %s), want value: %v, scale: %s", got.String(), got.Value(), got.Scale(), tt.wantValue, tt.wantScale)
			}
		})
	}
}

func TestMagnitude_String(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		scale string
		want  string
	}{
		{
			name:  "Magnitude with known scale",
			value: 5.5,
			scale: event.MagnitudeScaleMw,
			want:  "5.5 mw",
		},
		{
			name:  "Magnitude with unknown scale",
			value: 4.0,
			scale: "invalid_scale",
			want:  "4.0 unknown",
		},
		{
			name:  "Magnitude with zero value",
			value: 0.0,
			scale: event.MagnitudeScaleMl,
			want:  "0.0 ml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := event.NewMagnitude(tt.value, tt.scale)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := m.String()
			if got != tt.want {
				assert.Failf(t, "String() mismatch", "got: %s, want: %s", got, tt.want)
			}
		})
	}
}

func TestMagnitude_Value(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		scale string
		want  float64
	}{
		{
			name:  "Magnitude with valid value",
			value: 5.5,
			scale: event.MagnitudeScaleMw,
			want:  5.5,
		},
		{
			name:  "Magnitude with zero value",
			value: 0.0,
			scale: event.MagnitudeScaleMl,
			want:  0.0,
		},
		{
			name:  "Magnitude with negative value",
			value: -1.0,
			scale: event.MagnitudeScaleMb,
			want:  -1.0,
		},
		{
			name:  "Magnitude with maximum value",
			value: 10.0,
			scale: event.MagnitudeScaleMs,
			want:  10.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := event.NewMagnitude(tt.value, tt.scale)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := m.Value()
			if got != tt.want {
				assert.Failf(t, "Value() mismatch", "got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestMagnitude_Scale(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		scale string
		want  string
	}{
		{
			name:  "Magnitude with known scale",
			value: 5.5,
			scale: event.MagnitudeScaleMw,
			want:  event.MagnitudeScaleMw,
		},
		{
			name:  "Magnitude with unknown scale",
			value: 4.0,
			scale: "invalid_scale",
			want:  event.MagnitudeScaleUnknown,
		},
		{
			name:  "Magnitude with another known scale",
			value: 3.0,
			scale: event.MagnitudeScaleMl,
			want:  event.MagnitudeScaleMl,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := event.NewMagnitude(tt.value, tt.scale)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := m.Scale()
			if got != tt.want {
				assert.Failf(t, "Scale() mismatch", "got: %s, want: %s", got, tt.want)
			}
		})
	}
}

func TestMagnitude_IsKnown(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		value float64
		scale string
		want  bool
	}{
		{
			name:  "Known scale Mw returns true",
			value: 5.5,
			scale: event.MagnitudeScaleMw,
			want:  true,
		},
		{
			name:  "Known scale ml returns true",
			value: 3.0,
			scale: event.MagnitudeScaleMl,
			want:  true,
		},
		{
			name:  "Known scale mb returns true",
			value: 4.0,
			scale: event.MagnitudeScaleMb,
			want:  true,
		},
		{
			name:  "Unknown scale returns false",
			value: 4.5,
			scale: "invalid",
			want:  false,
		},
		{
			name:  "Empty scale returns false",
			value: 5.0,
			scale: "",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := event.NewMagnitude(tt.value, tt.scale)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			got := m.IsKnown()
			if got != tt.want {
				assert.Failf(t, "IsKnown() mismatch", "got: %v, want: %v", got, tt.want)
			}
		})
	}
}
