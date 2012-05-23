package polish

import (
  "strings"
  "fmt"
  "strconv"
  "reflect"
  "math"
  "runtime/debug"
)

type Error struct {
  ErrorString string

  // Stack trace where the error occurred, if available
  Stack []byte
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
  parse_order []Type
}

type Type int
const(
  Integer Type = iota
  Float
  String
)

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
  var val reflect.Value
  for _, v := range c.parse_order {
    switch v {
    case Integer:
      ival, e := strconv.Atoi(term)
      if e == nil {
        val = reflect.ValueOf(ival)
      }

    case Float:
      fval, e := strconv.ParseFloat(term, 64)
      if e == nil {
        val = reflect.ValueOf(fval)
      }

    case String:
      val = reflect.ValueOf(term)

    default:
      return nil, &Error{fmt.Sprintf("Unknown polish.Value: %v", v), nil}
    }
    if val != (reflect.Value{}) {
      break
    }
  }
  if val == (reflect.Value{}) {
    return nil, &Error{fmt.Sprintf("Unable to parse term: '%s'", term), nil}
  }
  vs = append(vs, val)
  return
}

// Evaluates a Polish notation expression using functions and values that have
// been specified using AddFunc and SetValue.
// Constants are interpreted as int if possible, otherwise float64.
func (c *Context) Eval(expression string) (vs []reflect.Value, err error) {
  defer func() {
    if r := recover(); r != nil {
      var local_err Error
      if e, ok := r.(error); ok {
        local_err.ErrorString = fmt.Sprintf("Failed to evaluate (%s): %s.", expression, e.Error())
      } else {
        local_err.ErrorString = fmt.Sprintf("Failed to evaluate (%s): %v.", expression, r)
      }
      local_err.Stack = debug.Stack()
      err = &local_err
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
    return &Error{fmt.Sprintf("Tried to add a %v instead of a function.", typ), nil}
  }
  if _, ok := c.funcs[name]; ok {
    return &Error{fmt.Sprintf("Tried to add the function '%s' more than once.", name), nil}
  }
  if _, ok := c.vals[name]; ok {
    return &Error{fmt.Sprintf("Tried to give the name '%s' to a function and a value.", name), nil}
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
    return &Error{fmt.Sprintf("Tried to give the name '%s' to a function and a value.", name), nil}
  }
  c.vals[name] = reflect.ValueOf(v)
  return nil
}

// Sets the order in which to attempt to parse terms.  The default order is
// Integer, Float, String.  You may want to specify that the order should be
// Float, String, for example, if you always want to deal with floating points
// without having to always specify a decimal point.
// String can parse anything, so if it comes before either Integer or Float
// then nothing will ever be parsed as those Types.
func (c *Context) SetParseOrder(types ...Type) {
  c.parse_order = types
}

// Makes a new Context with no functions or values.
func MakeContext() *Context {
  return &Context{
    funcs: make(map[string]function),
    vals:  make(map[string]reflect.Value),
    parse_order: []Type{Integer, Float, String},
  }
}

// Adds some basic boolean operators
//   Functions: && (logical and)
//              || (logical or)
//              ^^ (logical xor)
//              !  (logical not)
//   Constants: pi e
func AddBooleanContext(c *Context) {
  c.AddFunc("&&", func(a, b bool) bool { return a && b })
  c.AddFunc("||", func(a, b bool) bool { return a || b })
  c.AddFunc("^^", func(a, b bool) bool { return (a && !b) || (!a && b) })
  c.AddFunc("!", func(a bool) bool { return !a })
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
  c.AddFunc("abs", math.Abs)
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
  c.AddFunc("abs", func(a int) int { if a < 0 { return -a }; return a })
  c.AddFunc("<", func(a, b int) bool { return a < b })
  c.AddFunc("<=", func(a, b int) bool { return a <= b })
  c.AddFunc(">", func(a, b int) bool { return a > b })
  c.AddFunc(">=", func(a, b int) bool { return a >= b })
  c.AddFunc("==", func(a, b int) bool { return a == b })
}
