package fft_test

import (
  . "gospec"
  "gospec"
  "math"
  "polish"
)

func Float64ContextSpec(c gospec.Context) {
  c.Specify("Float64 context works properly.", func() {
    context := polish.MakeContext()
    polish.AddFloat64MathContext(context)
    v1 := math.E * math.Pi * math.Exp(1.23456 - math.Log10(77))
    res,err := context.Eval("* e * pi ^ e - 1.23456 log10 77.0")
    c.Assume(err, Equals, nil)
    c.Expect(res.Float(), IsWithin(1e-9), v1)
    res,err = context.Eval("< e pi")
    c.Assume(err, Equals, nil)
    c.Expect(res.Bool(), Equals, true)
  })
}

func IntContextSpec(c gospec.Context) {
  c.Specify("Int context works properly.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    v1 := (3 * 3 * 3) / (2 * 2 * 2 * 2) - (6 * 6)
    v2 := (5 * 5 * 5 * 5 * 5)
    res,err := context.Eval("- / ^ 3 3 ^ 2 4 ^ 6 2")
    c.Assume(err, Equals, nil)
    c.Expect(int(res.Int()), Equals, v1)
    res,err = context.Eval("^ 5 5")
    c.Assume(err, Equals, nil)
    c.Expect(int(res.Int()), Equals, v2)
    res,err = context.Eval("< - / ^ 3 3 ^ 2 4 ^ 6 2 ^ 5 5")
    c.Assume(err, Equals, nil)
    c.Expect(res.Bool(), Equals, v1 < v2)
  })
}
