package fft_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(Float64ContextSpec)
  r.AddSpec(IntContextSpec)
  gospec.MainGoTest(r, t)
}

