// Package gofp provides Functional-Programming like way
// to use Go.
package gofp

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
)

// Pipeline is a single-direction channel.
type Pipeline <-chan interface{}

// New creates a new Pipeline instances.
func New(f func(ch chan<- interface{})) Pipeline {
	out := make(chan interface{}, 1)
	go func() {
		defer close(out)
		f(out)
	}()
	return out
}

// ForEach creates a new Pipeline instances.
func ForEach(vs ...interface{}) Pipeline {
	return New(func(ch chan<- interface{}) {
		for _, v := range vs {
			ch <- v
		}
	})
}

// FromArray new Pipeline from any array or slice.
func FromArray(array interface{}) Pipeline {
	at := reflect.TypeOf(array)
	av := reflect.ValueOf(array)
	if at.Kind() != reflect.Array && at.Kind() != reflect.Slice {
		panic("need slice or array")
	}
	return New(func(ch chan<- interface{}) {
		for i := 0; i < av.Len(); i++ {
			ch <- av.Index(i).Interface()
		}
	})
}

// Range returns a new Pipeline which contains
// from start to end integer values.
func Range(init int, r ...int) Pipeline {
	switch len(r) {
	case 0:
		return RangeStep(0, init)
	case 1:
		return RangeStep(init, r[0])
	case 2:
		return RangeStep(init, r[0], r[1])
	default:
		return nil
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// RangeStep returns a new Pipeline which contains
// from start to end by step integer values.
func RangeStep(start, end int, steps ...int) Pipeline {
	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}
	t := end - start
	s := step * (t / abs(t))
	return New(func(out chan<- interface{}) {
		for i := 0; abs(i) < abs(t); i += s {
			out <- start + i
		}
	})
}

// Lines reads contents line by line from reader and passes into pipeline.
func Lines(r io.Reader) Pipeline {
	return scanReader(r, bufio.ScanLines)
}

// Words reads contents word by word from reader and passes into pipeline.
func Words(r io.Reader) Pipeline {
	return scanReader(r, bufio.ScanWords)
}

func scanReader(r io.Reader, split bufio.SplitFunc) Pipeline {
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	return New(func(out chan<- interface{}) {
		for scanner.Scan() {
			out <- scanner.Text()
		}
	})
}

// TakeAll returns all values in Pipeline.
func (pl Pipeline) TakeAll() []interface{} {
	var values []interface{}
	for v := range pl {
		values = append(values, v)
	}
	return values
}

// Take takes the first `n` elements from Pipeline.
func (pl Pipeline) Take(n int) []interface{} {
	if n <= 0 {
		return nil
	}
	var values []interface{}
	for v := range pl {
		values = append(values, v)
		n--
		if n <= 0 {
			break
		}
	}
	return values
}

// First takes the first element in Pipeline and returns.
func (pl Pipeline) First() interface{} {
	for v := range pl {
		return v
	}
	return nil
}

// Drop ignores the first n elements in Pipeline and
// returns itself.
func (pl Pipeline) Drop(n int) Pipeline {
	for i := 0; i < n; i++ {
		<-pl
	}
	return pl
}

// DropAll drops all values in Pipeline.
func (pl Pipeline) DropAll() {
	for _ = range pl {
	}
}

// Map passes each element in Pipeline into MapFunc.
func (pl Pipeline) Map(f interface{}) Pipeline {
	var mf MapFunc
	switch ft := f.(type) {
	case func(interface{}) interface{}:
		mf = MapFunc(ft)
	case MapFunc:
		mf = ft
	case Func:
		mf = ft.ToMapFunc()
	default:
		mf = NewFunc(f).ToMapFunc()
	}
	return New(func(out chan<- interface{}) {
		for i := range pl {
			out <- mf.Map(i)
		}
	})
}

// Filter drops all the invalid elements in Pipeline.
func (pl Pipeline) Filter(f interface{}) Pipeline {
	var ff FilterFunc
	switch ft := f.(type) {
	case func(interface{}) bool:
		ff = FilterFunc(ft)
	case FilterFunc:
		ff = ft
	case Func:
		ff = ft.ToFilterFunc()
	default:
		ff = NewFunc(f).ToFilterFunc()
	}
	return New(func(out chan<- interface{}) {
		for i := range pl {
			if ff.Filter(i) {
				out <- i
			}
		}
	})
}

// Reduce reduces all elements in Pipeline to a final result.
func (pl Pipeline) Reduce(f, init interface{}) interface{} {
	var rf ReduceFunc
	switch ft := f.(type) {
	case func(interface{}, interface{}) interface{}:
		rf = ReduceFunc(rf)
	case ReduceFunc:
		rf = ft
	case Func:
		rf = ft.ToReduceFunc()
	default:
		rf = NewFunc(f).ToReduceFunc()
	}
	result := init
	for i := range pl {
		result = rf.Reduce(i, result)
	}
	return result
}

// Maybe type
type Maybe struct {
	v interface{}
}

// Nothing is a nil Maybe value.
var Nothing = &Maybe{nil}

// Just maybe value
func Just(v interface{}) *Maybe {
	if v == nil {
		return Nothing
	}
	return &Maybe{v: v}
}

// Map applies func to value in Maybe context
func (m *Maybe) Map(f interface{}) *Maybe {
	if m == Nothing {
		return Nothing
	}
	var mf MapFunc
	switch ft := f.(type) {
	case func(interface{}) interface{}:
		mf = MapFunc(ft)
	case MapFunc:
		mf = ft
	case Func:
		mf = ft.ToMapFunc()
	default:
		mf = NewFunc(f).ToMapFunc()
	}
	return Just(mf.Map(m.v))
}

// Join removes one leve of Just.
func (m *Maybe) Join() *Maybe {
	if m == Nothing {
		return Nothing
	}
	in, ok := m.v.(*Maybe)
	if ok {
		return in
	}
	return m
}

func (m *Maybe) String() string {
	if m == Nothing {
		return "Nothing"
	}

	return fmt.Sprintf("Just %v", m.v)
}

// Func type
type Func func(...interface{}) reflect.Value

// NewFunc create a new Func.
func NewFunc(f interface{}) Func {
	return func(args ...interface{}) reflect.Value {
		fv := reflect.ValueOf(f)
		var vargs []reflect.Value
		for _, arg := range args {
			vargs = append(vargs, reflect.ValueOf(arg))
		}
		results := fv.Call(vargs)
		return results[0]
	}
}

// Call applies args to Func
func (f Func) Call(args ...interface{}) reflect.Value {
	return f(args...)
}

// Curry returns a new Func with a default first arg
func (f Func) Curry(vs ...interface{}) Func {
	return func(args ...interface{}) reflect.Value {
		return f.Call(append(vs, args...)...)
	}
}

// Flip args
func (f Func) Flip() Func {
	return func(args ...interface{}) reflect.Value {
		return f.Call(append(args[1:], args[0])...)
	}
}

// FlipCurry is a fast method to call Flip an Curry, as this
// situiation is very common in daily use.
func (f Func) FlipCurry(vs ...interface{}) Func {
	return f.Flip().Curry(vs...)
}

// MapFunc type
type MapFunc func(interface{}) interface{}

// Map easy method
func (mf MapFunc) Map(v interface{}) interface{} {
	return mf(v)
}

// LiftMaybe lifts normal func processing Maybe type
func (mf MapFunc) LiftMaybe() MapFunc {
	return NewFunc(func(m *Maybe) *Maybe {
		return m.Map(mf).Join()
	}).ToMapFunc()
}

// ToMapFunc converts Func to MapFunc
func (f Func) ToMapFunc() MapFunc {
	return func(v interface{}) interface{} {
		return f.Call(v).Interface()
	}
}

// FilterFunc type
type FilterFunc func(interface{}) bool

// Filter easy method
func (ff FilterFunc) Filter(v interface{}) bool {
	return ff(v)
}

// Not reverses FilterFunc
func (ff FilterFunc) Not() FilterFunc {
	return func(v interface{}) bool {
		return !ff.Filter(v)
	}
}

// ToFilterFunc converts Func to FilterFunc
func (f Func) ToFilterFunc() FilterFunc {
	return func(v interface{}) bool {
		return f.Call(v).Bool()
	}
}

// ReduceFunc type
type ReduceFunc func(v1, v2 interface{}) interface{}

// Reduce easy method
func (rf ReduceFunc) Reduce(v1, v2 interface{}) interface{} {
	return rf(v1, v2)
}

// ToReduceFunc converts Func to ReduceFunc
func (f Func) ToReduceFunc() ReduceFunc {
	return func(v1, v2 interface{}) interface{} {
		return f.Call(v1, v2).Interface()
	}
}

// NothingFilter is a FilterFunc to fuck all Nothing value
var NothingFilter = func(m *Maybe) bool {
	return m != Nothing
}
