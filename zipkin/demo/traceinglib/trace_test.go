package traceinglib

import (
	"testing"
)

func TestInit(t *testing.T) {
	f := InitTracer()
	defer f()
}
