package dsp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFollow(t *testing.T) {
	rise, fall := 1.0, 2.0
	follow := &Follow{}
	require.Equal(t, 0.0, follow.Tick(0.0, rise, fall))
	require.Equal(t, 0.495, follow.Tick(0.5, rise, fall))
	require.Equal(t, 0.0495, follow.Tick(0.0, rise, fall))
	require.Equal(t, 0.495495, follow.Tick(-0.5, rise, fall))
	require.Equal(t, 0.0495495, follow.Tick(0.0, rise, fall))
	require.Equal(t, 0.0049549500000000005, follow.Tick(0.0, rise, fall))
	require.Equal(t, 0.0004954950000000001, follow.Tick(0.0, rise, fall))
}
