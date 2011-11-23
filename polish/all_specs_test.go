package polish_test

import (
  "gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(Float64ContextSpec)
  r.AddSpec(IntContextSpec)
  r.AddSpec(MultiValueReturnSpec)
  gospec.MainGoTest(r, t)
}
