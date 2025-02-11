package apmkratoslog

import (
	"testing"

	"go.elastic.co/apm/v2"
)

func TestNewLogHelper(t *testing.T) {
	if false {
		var _ apm.Logger = &logHelper{}
	}
}

func TestNewApmLogger(t *testing.T) {
	if false {
		var _ apm.Logger = &apmLogger{}
	}
}
