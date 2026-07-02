// Package dslbuilder - Typed action helpers.
//
// The raw ActionFunc signature (func([]interface{}) (interface{}, error))
// forces type assertions in every action. These helpers keep that API for
// compatibility but remove the boilerplate:
//
//	dsl.Action("add", dslbuilder.Action3(func(left int, _ string, right int) (int, error) {
//	    return left + right, nil
//	}))
//
//	dsl.Action("number", dslbuilder.Action1(func(text string) (int, error) {
//	    return strconv.Atoi(text)
//	}))
//
// Arguments are converted with ToInt/ToFloat/ToString semantics: numeric
// kinds convert between each other, and strings parse into numbers. A
// conversion failure returns a descriptive error instead of panicking.
package dslbuilder

import (
	"fmt"
	"strconv"
)

// Action1 adapts a typed 1-argument function to an ActionFunc.
func Action1[A, R any](fn func(A) (R, error)) ActionFunc {
	return func(args []interface{}) (interface{}, error) {
		if err := wantArgs(args, 1); err != nil {
			return nil, err
		}
		a, err := convertArg[A](args, 0)
		if err != nil {
			return nil, err
		}
		return fn(a)
	}
}

// Action2 adapts a typed 2-argument function to an ActionFunc.
func Action2[A, B, R any](fn func(A, B) (R, error)) ActionFunc {
	return func(args []interface{}) (interface{}, error) {
		if err := wantArgs(args, 2); err != nil {
			return nil, err
		}
		a, err := convertArg[A](args, 0)
		if err != nil {
			return nil, err
		}
		b, err := convertArg[B](args, 1)
		if err != nil {
			return nil, err
		}
		return fn(a, b)
	}
}

// Action3 adapts a typed 3-argument function to an ActionFunc.
// Typical for infix operators: [left, operator, right].
func Action3[A, B, C, R any](fn func(A, B, C) (R, error)) ActionFunc {
	return func(args []interface{}) (interface{}, error) {
		if err := wantArgs(args, 3); err != nil {
			return nil, err
		}
		a, err := convertArg[A](args, 0)
		if err != nil {
			return nil, err
		}
		b, err := convertArg[B](args, 1)
		if err != nil {
			return nil, err
		}
		c, err := convertArg[C](args, 2)
		if err != nil {
			return nil, err
		}
		return fn(a, b, c)
	}
}

func wantArgs(args []interface{}, n int) error {
	if len(args) != n {
		return fmt.Errorf("action expected %d arguments, got %d", n, len(args))
	}
	return nil
}

func convertArg[T any](args []interface{}, i int) (T, error) {
	var zero T
	v := args[i]

	// Direct type match first.
	if t, ok := v.(T); ok {
		return t, nil
	}

	// Fall back to flexible conversion via the Args helpers.
	var converted interface{}
	var err error
	switch any(zero).(type) {
	case int:
		converted, err = toInt(v)
	case int64:
		var n int
		n, err = toInt(v)
		converted = int64(n)
	case float64:
		converted, err = toFloat(v)
	case string:
		converted = fmt.Sprint(v)
	default:
		return zero, fmt.Errorf("argument %d: cannot convert %T to %T", i, v, zero)
	}
	if err != nil {
		return zero, fmt.Errorf("argument %d: %w", i, err)
	}
	return converted.(T), nil
}

// Args wraps the raw argument slice of an ActionFunc with typed accessors:
//
//	dsl.Action("add", func(raw []interface{}) (interface{}, error) {
//	    args := dslbuilder.Args(raw)
//	    return args.Int(0) + args.Int(2), nil
//	})
//
// Accessors are best-effort: on conversion failure they return the zero
// value. Use the To* functions when you need the error.
type Args []interface{}

// Len returns the number of arguments.
func (a Args) Len() int { return len(a) }

// Get returns the raw argument at i, or nil if out of range.
func (a Args) Get(i int) interface{} {
	if i < 0 || i >= len(a) {
		return nil
	}
	return a[i]
}

// String returns the argument at i as a string (fmt.Sprint for non-strings).
func (a Args) String(i int) string {
	v := a.Get(i)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// Int returns the argument at i converted to int (0 on failure).
func (a Args) Int(i int) int {
	n, _ := toInt(a.Get(i))
	return n
}

// Float returns the argument at i converted to float64 (0 on failure).
func (a Args) Float(i int) float64 {
	f, _ := toFloat(a.Get(i))
	return f
}

// toInt converts numeric kinds and numeric strings to int.
func toInt(v interface{}) (int, error) {
	switch n := v.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case float64:
		return int(n), nil
	case string:
		i, err := strconv.Atoi(n)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to int", n)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// toFloat converts numeric kinds and numeric strings to float64.
func toFloat(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot convert %q to float", n)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", v)
	}
}
