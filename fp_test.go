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

func TestFirst(t *testing.T) {
	pl := Range(1, 2)
	first := pl.First()
	if first.(int) != 1 {
		t.Errorf("want %d got %d", 1, first)
	}
	first = pl.First()
	if first != nil {
		t.Errorf("want %v got %v", nil, first)
	}
}

func TestDrop(t *testing.T) {
	pl := Range(0, 10)
	pl.Drop(5)
	rest := pl.TakeAll()
	if len(rest) != 5 {
		t.Errorf("want %d got %d", 5, len(rest))
	}
}

func TestRange(t *testing.T) {
	all := Range(0, 10).TakeAll()
	for i, v := range all {
		if v.(int) != i {
			t.Errorf("want %d got %d", i, v)
		}
	}
}

func TestRangeStep(t *testing.T) {
	all := RangeStep(0, 10, 1).TakeAll()
	for i, v := range all {
		if v.(int) != i {
			t.Errorf("want %d got %d", i, v)
		}
	}

	all = RangeStep(10, 0, -1).TakeAll()
	for i, v := range all {
		if v.(int) != 10-i {
			t.Errorf("want %d got %d", 10-i, v)
		}
	}
}

func TestMap(t *testing.T) {
	values := Range(1, 5).Map(func(i int) int {
		return i + 1
	}).Map(func(i int) int {
		return i * 2
	}).TakeAll()

	for i, v := range values {
		if v.(int) != (i+2)*2 {
			t.Errorf("want %d got %d", (i+2)*2, v)
		}
	}
}

func TestFilter(t *testing.T) {
	values := Range(1, 6).Filter(func(i int) bool {
		return i%2 == 0
	}).TakeAll()
	for _, v := range values {
		if v.(int)%2 != 0 {
			t.Errorf("%d", v)
		}
	}
}

func TestReduce(t *testing.T) {
	result := ForEach(1, 2, 3, 4, 5).Reduce(func(i, j int) int {
		return i + j
	}, 0)
	if result.(int) != 15 {
		t.Error("want %d got %d", 15, result)
	}
}

func TestMaybe(t *testing.T) {
	inc := func(i int) int { return i + 1 }
	if res := Nothing.Map(inc); res != Nothing {
		t.Errorf("want Nothing got %s", res)
	}

	if res := Just(1).Map(inc); res.v != Just(2).v {
		t.Errorf("want %s got %s", Just(2), res)
	}
}

func BenchmarkMap(b *testing.B) {
	pl := ForEach(1, 2, 3, 4)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Map(func(i int) int {
			return i + 1
		}).Map(func(i int) int {
			return i * 2
		}).TakeAll()
	}
}

func BenchmarkFilter(b *testing.B) {
	pl := ForEach(1, 2, 3, 4, 5)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Filter(func(i int) bool {
			return i%2 == 0
		}).TakeAll()
	}
}

func BenchmarkReduce(b *testing.B) {
	pl := ForEach(1, 2, 3, 4, 5)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pl.Reduce(func(i, j int) int {
			return i + j
		}, 0)
	}
}
