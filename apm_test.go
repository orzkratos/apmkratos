package apmkratos

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.elastic.co/apm/v2"
)

// TestGetApmAgentVersion tests APM agent version access
// 测试获取 APM agent 版本
func TestGetApmAgentVersion(t *testing.T) {
	t.Log(GetApmAgentVersion())
}

// TestCheckApmAgentVersion tests APM version alignment check
// 测试 APM 版本一致性检查
func TestCheckApmAgentVersion(t *testing.T) {
	require.True(t, CheckApmAgentVersion(apm.AgentVersion))
}
