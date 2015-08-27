// Package gofp provides Functional-Programming like way
// to use Go.
package gofp

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
		n--
		values = append(values, v)
		if n <= 0 {
			break
		}
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

func Range(start, end int) Pipeline {
	return RangeStep(start, end, 1)
}

func RangeStep(start, end, step int) Pipeline {
	out := make(chan interface{}, 1)
	go func() {
		defer close(out)
		for i := start; i < end; i += step {
			out <- i
		}
	}()
	return out
}
