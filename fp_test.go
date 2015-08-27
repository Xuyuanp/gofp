package gofp

import "testing"

func TestTake(t *testing.T) {
	pl := Range(1, 6)
	values := pl.Take(0)
	if values != nil {
		t.Errorf("want %s got %s", nil, values)
	}
	values = pl.Take(1)
	if values == nil || len(values) == 0 {
		t.Errorf("take 1 got empty array")
	}
	if values[0].(int) != 1 {
		t.Errorf("first element not equal to %d", 1)
	}

	values = pl.Take(5)
	if len(values) != 4 {
		t.Errorf("want %d got %d", 4, len(values))
	}
}

func TestMap(t *testing.T) {
	values := Range(1, 5).Map(
		func(v interface{}) interface{} {
			return v.(int) + 1
		},
		func(v interface{}) interface{} {
			return v.(int) * 2
		}).Values()

	for i, v := range values {
		if v.(int) != (i+2)*2 {
			t.Errorf("want %d got %d", (i+2)*2, v)
		}
	}
}

func TestFilter(t *testing.T) {
	values := Range(1, 6).Filter(
		func(v interface{}) bool {
			return v.(int)%2 == 0
		}).
		Values()
	for _, v := range values {
		if v.(int)%2 != 0 {
			t.Errorf("%d", v)
		}
	}
}

func TestReduce(t *testing.T) {
	result := NewPipeline(1, 2, 3, 4, 5).Reduce(
		func(v1, v2 interface{}) interface{} {
			return v1.(int) + v2.(int)
		}, 0)
	if result.(int) != 15 {
		t.Error("want %d got %d", 15, result)
	}
}

func BenchmarkMap(b *testing.B) {
	pl := NewPipeline(1, 2, 3, 4)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Map(
			func(v interface{}) interface{} {
				return v.(int) + 1
			},
			func(v interface{}) interface{} {
				return v.(int) * 2
			}).Values()
	}
}

func BenchmarkFilter(b *testing.B) {
	pl := NewPipeline(1, 2, 3, 4, 5)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Filter(
			func(v interface{}) bool {
				return v.(int)%2 == 0
			}).
			Values()
	}
}

func BenchmarkReduce(b *testing.B) {
	pl := NewPipeline(1, 2, 3, 4, 5)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Reduce(
			func(v1, v2 interface{}) interface{} {
				return v1.(int) + v2.(int)
			}, 0)
	}
}
