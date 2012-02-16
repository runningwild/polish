package polish_test

import (
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/orfjackal/gospec/src/gospec"
  "math"
  "github.com/runningwild/polish"
)

func Float64ContextSpec(c gospec.Context) {
  c.Specify("Float64 context works properly.", func() {
    context := polish.MakeContext()
    polish.AddFloat64MathContext(context)
    v1 := math.E * math.Pi * math.Exp(1.23456-math.Log10(77))
    res, err := context.Eval("* e * pi ^ e - 1.23456 log10 77.0")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(res[0].Float(), IsWithin(1e-9), v1)
    res, err = context.Eval("< e pi")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(res[0].Bool(), Equals, true)
  })
}

func IntContextSpec(c gospec.Context) {
  c.Specify("Int context works properly.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    v1 := (3*3*3)/(2*2*2*2) - (6 * 6)
    v2 := (5 * 5 * 5 * 5 * 5)
    res, err := context.Eval("- / ^ 3 3 ^ 2 4 ^ 6 2")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(int(res[0].Int()), Equals, v1)
    res, err = context.Eval("^ 5 5")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(int(res[0].Int()), Equals, v2)
    res, err = context.Eval("< - / ^ 3 3 ^ 2 4 ^ 6 2 ^ 5 5")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(res[0].Bool(), Equals, v1 < v2)
  })
}

func MultiValueReturnSpec(c gospec.Context) {
  c.Specify("Functions with zero or more than one return values work.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    rev3 := func(a, b, c int) (int, int, int) {
      return c, b, a
    }
    context.AddFunc("rev3", rev3)
    rev5 := func(a, b, c, d, e int) (int, int, int, int, int) {
      return e, d, c, b, a
    }
    context.AddFunc("rev5", rev5)

    res, err := context.Eval("- - - - rev5 rev3 1 2 rev3 4 5 6")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    // - - - - rev5 rev3 1 2 rev3 4 5 6
    // - - - - rev5 rev3 1 2 6 5 4
    // - - - - rev5 6 2 1 5 4
    // - - - - 4 5 1 2 6
    // - - - -1 1 2 6
    // - - -2 2 6
    // - -4 6
    // -10
    c.Expect(int(res[0].Int()), Equals, -10)
  })
}

func ErrorSpec(c gospec.Context) {
  c.Specify("Type-mismatch panics are caught and returned as errors.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    _, err := context.Eval("+ 1.0 2.0")
    c.Assume(err.Error(), Not(Equals), nil)
  })
  c.Specify("Panics from inside functions are caught and returned as errors.", func() {
    context := polish.MakeContext()
    context.AddFunc("panic", func() { panic("rawr") })
    _, err := context.Eval("panic")
    c.Assume(err.Error(), Not(Equals), nil)
  })
}

func NumRemainingValuesSpec(c gospec.Context) {
  c.Specify("Can handle any number of terms remaining after evaluation.", func() {
    context := polish.MakeContext()
    context.AddFunc("makeZero", func() {})
    context.AddFunc("makeTwo", func() (int, int) { return 1, 2 })
    res, err := context.Eval("makeTwo")
    c.Assume(len(res), Equals, 2)
    c.Assume(err, Equals, nil)
    res, err = context.Eval("makeZero")
    c.Assume(len(res), Equals, 0)
    c.Assume(err, Equals, nil)
  })
}

func ParsingSpec(c gospec.Context) {
  c.Specify("Whitespace is parsed properly.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    res, err := context.Eval("    +           1                      3")
    c.Assume(len(res), Equals, 1)
    c.Assume(err, Equals, nil)
    c.Expect(int(res[0].Int()), Equals, 4)
  })
}

func IntOperatorSpec(c gospec.Context) {
  c.Specify("All standard int operators parse.", func() {
    context := polish.MakeContext()
    polish.AddIntMathContext(context)
    c.Specify("+ works", func() {
      _, err := context.Eval("+ 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("- works", func() {
      _, err := context.Eval("- 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("* works", func() {
      _, err := context.Eval("* 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("/ works", func() {
      _, err := context.Eval("/ 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("< works", func() {
      _, err := context.Eval("< 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("<= works", func() {
      _, err := context.Eval("<= 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("> works", func() {
      _, err := context.Eval("> 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify(">= works", func() {
      _, err := context.Eval(">= 1 2")
      c.Assume(err, Equals, nil)
    })
    c.Specify("== works", func() {
      _, err := context.Eval("== 1 2")
      c.Assume(err, Equals, nil)
    })
  })
}
