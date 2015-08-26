// Package gofp provides Functional-Programming like way
// to use Go.
package gofp

// Pipeline is a single-direction channel.
type Pipeline <-chan interface{}

// NewPipeline create a new Pipeline instances.
func NewPipeline(vs ...interface{}) Pipeline {
	pl := make(chan interface{}, 1)
	go func() {
		defer close(pl)
		for _, v := range vs {
			pl <- v
		}
	}()
	return pl
}

// Values returns all values in Pipeline.
func (pl Pipeline) Values() []interface{} {
	var values []interface{}
	for v := range pl {
		values = append(values, v)
	}
	return values
}

// MapFunc functions processes each element in Pipeline.
type MapFunc func(interface{}) interface{}

// Map passes each element in Pipeline into MapFunc.
func (pl Pipeline) Map(fs ...MapFunc) Pipeline {
	out := make(chan interface{}, 1)
	go func() {
		defer close(out)
		for i := range pl {
			for _, f := range fs {
				i = f(i)
			}
			out <- i
		}
	}()
	return out
}

// FilterFunc ignores all the elements which returns false.
type FilterFunc func(interface{}) bool

// Filter drops all the invalid elements in Pipeline.
func (pl Pipeline) Filter(f FilterFunc) Pipeline {
	out := make(chan interface{}, 1)
	go func() {
		defer close(out)
		for i := range pl {
			if f(i) {
				out <- i
			}
		}
	}()
	return out
}

// ReduceFunc reduces 2 elements into a single one.
type ReduceFunc func(interface{}, interface{}) interface{}

// Reduce reduces all elements in Pipeline to a final result.
func (pl Pipeline) Reduce(f ReduceFunc, init interface{}) interface{} {
	result := init
	for i := range pl {
		result = f(i, result)
	}
	return result
}
