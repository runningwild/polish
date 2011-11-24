package polish

import (
  "strings"
  "fmt"
  "strconv"
  "reflect"
  "math"
)

type Error struct {
  ErrorString string
}

func (e *Error) Error() string {
  return e.ErrorString
}

type function struct {
  // An arbitrary function
  f reflect.Value

  // The number of input values for the above function
  num int
}

// A Context is used to evaluate Polish notation expressions.  The Context
// provides functions and values that can be used in the expressions.  A basic
// math context might be created as follows:
//   c := MakeContext()
//   c.AddFunc("+", func(a,b float64) float64 { return a + b })
//   c.AddFunc("-", func(a,b float64) float64 { return a - b })
//   c.AddFunc("*", func(a,b float64) float64 { return a * b })
//   c.AddFunc("/", func(a,b float64) float64 { return a / b })
//   c.SetValue("pi", math.Pi)
//   c.SetValue("e", math.E)
// At which point expressions could be evaluated like so:
//   v,err = c.Eval("* 2.0 pi")
//   v.Float()  // Evaluates to 2 * math.Pi
//   v,err = c.Eval("* 3.0 - pi e ")
//   v.Float()  // Evaluates to 3 * (pi - e)
// Constants are interpreted as int if possible, otherwise float64.
type Context struct {
  funcs map[string]function
  vals  map[string]reflect.Value
  terms []string
}

func (c *Context) subEval() (vs []reflect.Value, err error) {
  term := c.terms[0]
  c.terms = c.terms[1:]
  if f, ok := c.funcs[term]; ok {
    var args []reflect.Value
    for len(args) < f.num {
      var results []reflect.Value
      results, err = c.subEval()
      if err != nil {
        return
      }
      for _, result := range results {
        args = append(args, result)
      }
    }
    var remaining []reflect.Value
    if len(args) > f.num {
      remaining = args[f.num:]
      args = args[0:f.num]
    }
    vs = f.f.Call(args)
    for _, v := range remaining {
      vs = append(vs, v)
    }
    return
  } else if val, ok := c.vals[term]; ok {
    vs = append(vs, val)
    return
  }
  fval, e := strconv.Atoi(term)
  if e != nil {
    ival, e := strconv.Atof64(term)
    if e != nil {
      err = e
      return
    }
    vs = append(vs, reflect.ValueOf(ival))
  } else {
    vs = append(vs, reflect.ValueOf(fval))
  }
  return
}

// Evaluates a Polish notation expression using functions and values that have
// been specified using AddFunc and SetValue.
// Constants are interpreted as int if possible, otherwise float64.
func (c *Context) Eval(expression string) (vs []reflect.Value, err error) {
  defer func() {
    if r := recover(); r != nil {
      if e, ok := r.(error); ok {
        err = &Error{fmt.Sprintf("Failed to evaluate: %s.", e.Error())}
      } else {
        err = &Error{fmt.Sprintf("Failed to evaluate: %v.", r)}
      }
    }
  }()
  raw_terms := strings.Split(expression, " ")
  c.terms = nil
  for _, term := range raw_terms {
    if len(term) > 0 {
      c.terms = append(c.terms, term)
    }
  }
  vs, err = c.subEval()
  if err != nil {
    return
  }
  return
}

// Adds a function that can be used in future calls to Eval.  Functions cannot
// be reassigned.
func (c *Context) AddFunc(name string, f interface{}) error {
  typ := reflect.TypeOf(f)
  if typ.Kind() != reflect.Func {
    return &Error{fmt.Sprintf("Tried to add a %v instead of a function.", typ)}
  }
  if _, ok := c.funcs[name]; ok {
    return &Error{fmt.Sprintf("Tried to add the function '%s' more than once.", name)}
  }
  if _, ok := c.vals[name]; ok {
    return &Error{fmt.Sprintf("Tried to give the name '%s' to a function and a value.", name)}
  }
  c.funcs[name] = function{
    f:   reflect.ValueOf(f),
    num: reflect.TypeOf(f).NumIn(),
  }
  return nil
}

// Sets a value that can be used in future calls to Eval.  Values can be
// reassigned
func (c *Context) SetValue(name string, v interface{}) error {
  if _, ok := c.funcs[name]; ok {
    return &Error{fmt.Sprintf("Tried to give the name '%s' to a function and a value.", name)}
  }
  c.vals[name] = reflect.ValueOf(v)
  return nil
}

// Makes a new Context with no functions or values.
func MakeContext() *Context {
  return &Context{
    funcs: make(map[string]function),
    vals:  make(map[string]reflect.Value),
  }
}

// Adds several operators and constants to the Context, all of which use float64
// for any numerical values.  
//   Functions: + - * / ^ ln log2 log10 < <= > >= ==
//   Constants: pi e
func AddFloat64MathContext(c *Context) {
  c.AddFunc("+", func(a, b float64) float64 { return a + b })
  c.AddFunc("-", func(a, b float64) float64 { return a - b })
  c.AddFunc("*", func(a, b float64) float64 { return a * b })
  c.AddFunc("/", func(a, b float64) float64 { return a / b })
  c.AddFunc("^", math.Pow)
  c.AddFunc("ln", math.Log)
  c.AddFunc("log2", math.Log2)
  c.AddFunc("log10", math.Log10)
  c.AddFunc("<", func(a, b float64) bool { return a < b })
  c.AddFunc("<=", func(a, b float64) bool { return a <= b })
  c.AddFunc(">", func(a, b float64) bool { return a > b })
  c.AddFunc(">=", func(a, b float64) bool { return a >= b })
  c.AddFunc("==", func(a, b float64) bool { return a == b })
  c.SetValue("pi", math.Pi)
  c.SetValue("e", math.E)
}

func iPow(base, exp int) int {
  if exp < 0 {
    panic("Cannot raise to a negative power when using integer exponentiation.")
  }
  if exp == 0 {
    return 1
  }
  return base * iPow(base, exp-1)
}

// Adds several operators to the Context, all of which use int for any numerical
// values.
//   Functions: + - * / ^ < <= > >= ==
func AddIntMathContext(c *Context) {
  c.AddFunc("+", func(a, b int) int { return a + b })
  c.AddFunc("-", func(a, b int) int { return a - b })
  c.AddFunc("*", func(a, b int) int { return a * b })
  c.AddFunc("/", func(a, b int) int { return a / b })
  c.AddFunc("^", iPow)
  c.AddFunc("<", func(a, b int) bool { return a < b })
  c.AddFunc("<=", func(a, b int) bool { return a <= b })
  c.AddFunc(">", func(a, b int) bool { return a > b })
  c.AddFunc(">=", func(a, b int) bool { return a >= b })
  c.AddFunc("==", func(a, b int) bool { return a == b })
}
