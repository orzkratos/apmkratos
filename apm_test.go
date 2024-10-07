package apmkratos

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.elastic.co/apm/v2"
)

func TestGetApmAgentVersion(t *testing.T) {
	t.Log(GetApmAgentVersion())
}

func TestCheckApmAgentVersion(t *testing.T) {
	require.True(t, CheckApmAgentVersion(apm.AgentVersion))
}
