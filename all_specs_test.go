package polish_test

import (
  "github.com/orfjackal/gospec/src/gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(Float64ContextSpec)
  r.AddSpec(Float64AndBooleanContextSpec)
  r.AddSpec(IntContextSpec)
  r.AddSpec(MultiValueReturnSpec)
  r.AddSpec(ErrorSpec)
  r.AddSpec(NumRemainingValuesSpec)
  r.AddSpec(ParsingSpec)
  r.AddSpec(IntOperatorSpec)
  gospec.MainGoTest(r, t)
}
