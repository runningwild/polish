package polish

import (
  "scanner"
  "fmt"
  "bytes"
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
  // An arbitrary function that must have exactly one return value
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
}

func (c *Context) subEval(scan *scanner.Scanner) (v reflect.Value, err error) {
  scan.Scan()
  token := scan.TokenText()
  if f, ok := c.funcs[token]; ok {
    var args []reflect.Value
    for i := 0; i < f.num; i++ {
      var result reflect.Value
      result, err = c.subEval(scan)
      if err != nil {
        return
      }
      args = append(args, result)
    }
    res := f.f.Call(args)
    v = res[0]
    return
  } else if val, ok := c.vals[token]; ok {
    v = val
    return
  }
  fval, e := strconv.Atoi(token)
  if e != nil {
    ival, e := strconv.Atof64(token)
    if e != nil {
      err = e
      return
    }
    v = reflect.ValueOf(ival)
  } else {
    v = reflect.ValueOf(fval)
  }
  return
}

// Evaluates a Polish notation expression using functions and values that have
// been specified using AddFunc and SetValue.
// Constants are interpreted as int if possible, otherwise float64.
func (c *Context) Eval(expression string) (v reflect.Value, err error) {
  defer func() {
    if r := recover(); r != nil {
      if e, ok := r.(error); ok {
        err = &Error{fmt.Sprintf("Failed to evaluate: %s.", e.Error())}
      } else {
        err = &Error{fmt.Sprintf("Failed to evaluate: %v.", r)}
      }
    }
  }()
  var scan scanner.Scanner
  scan.Init(bytes.NewBufferString(expression))
  return c.subEval(&scan)
}

// Adds a function that can be used in future calls to Eval.  f must have
// exactly one output.  Functions cannot be reassigned.
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