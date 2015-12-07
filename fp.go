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

// Range returns a new Pipeline which contains
// from start to end integer values.
func Range(start, end int) Pipeline {
	return RangeStep(start, end, 1)
}

// RangeStep returns a new Pipeline which contains
// from start to end by step integer values.
func RangeStep(start, end, step int) Pipeline {
	return New(func(out chan<- interface{}) {
		if step > 0 {
			for i := start; i < end; i += step {
				out <- i
			}
		} else {
			for i := start; i > end; i += step {
				out <- i
			}
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

// MapFunc functions processes each element in Pipeline.
type MapFunc struct {
	in  reflect.Type
	out reflect.Type
	f   func(interface{}) interface{}
}

// Map calls inner func in MapFunc.
func (f *MapFunc) Map(v interface{}) interface{} {
	return f.f(v)
}

func newMapFunc(f interface{}) *MapFunc {
	tf := reflect.TypeOf(f)
	vf := reflect.ValueOf(f)
	if tf.Kind() != reflect.Func {
		panic("not a function")
	}
	return &MapFunc{
		in:  tf.In(0),
		out: tf.Out(0),
		f: func(x interface{}) interface{} {
			out := vf.Call([]reflect.Value{reflect.ValueOf(x)})
			return out[0].Interface()
		},
	}
}

// Map passes each element in Pipeline into MapFunc.
func (pl Pipeline) Map(f interface{}) Pipeline {
	mf := newMapFunc(f)
	return New(func(out chan<- interface{}) {
		for i := range pl {
			out <- mf.Map(i)
		}
	})
}

// FilterFunc ignores all the elements which returns false.
type FilterFunc struct {
	in reflect.Type
	f  func(interface{}) bool
}

// Filter calls inner func of FilterFunc
func (f *FilterFunc) Filter(v interface{}) bool {
	return f.f(v)
}

func newFilterFunc(f interface{}) *FilterFunc {
	tf := reflect.TypeOf(f)
	vf := reflect.ValueOf(f)
	if tf.Kind() != reflect.Func {
		panic("not a function")
	}
	if tf.NumOut() != 1 || tf.Out(0).Kind() != reflect.Bool {
		panic("return type is not bool")
	}
	return &FilterFunc{
		in: tf.In(0),
		f: func(x interface{}) bool {
			out := vf.Call([]reflect.Value{reflect.ValueOf(x)})
			return out[0].Bool()
		},
	}
}

// Filter drops all the invalid elements in Pipeline.
func (pl Pipeline) Filter(f interface{}) Pipeline {
	ff := newFilterFunc(f)
	return New(func(out chan<- interface{}) {
		for i := range pl {
			if ff.Filter(i) {
				out <- i
			}
		}
	})
}

// ReduceFunc reduces 2 elements into a single one.
type ReduceFunc struct {
	in1 reflect.Type
	in2 reflect.Type
	out reflect.Type
	f   func(interface{}, interface{}) interface{}
}

// Reduce calls inner function of ReduceFunc.
func (f *ReduceFunc) Reduce(v1, v2 interface{}) interface{} {
	return f.f(v1, v2)
}

func newReduceFunc(f interface{}) *ReduceFunc {
	tf := reflect.TypeOf(f)
	vf := reflect.ValueOf(f)
	if tf.Kind() != reflect.Func {
		panic("not a function")
	}
	if tf.NumIn() != 2 {
		panic("require 2 params")
	}
	if tf.NumOut() != 1 {
		panic("require 1 return value")
	}
	if tf.In(1) != tf.Out(0) {
		panic("types of param 2 and return value doesn't match")
	}
	return &ReduceFunc{
		in1: tf.In(0),
		in2: tf.In(1),
		out: tf.Out(0),
		f: func(v1, v2 interface{}) interface{} {
			out := vf.Call([]reflect.Value{reflect.ValueOf(v1), reflect.ValueOf(v2)})
			return out[0].Interface()
		},
	}
}

// Reduce reduces all elements in Pipeline to a final result.
func (pl Pipeline) Reduce(f interface{}, init interface{}) interface{} {
	rf := newReduceFunc(f)
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
	mf := newMapFunc(f)
	if m == Nothing {
		return Nothing
	}
	return Just(mf.Map(m.v))
}

func (m *Maybe) String() string {
	if m == Nothing {
		return "Nothing"
	}

	return fmt.Sprintf("Just %v", m.v)
}
