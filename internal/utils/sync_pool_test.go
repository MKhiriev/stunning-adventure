package utils

import "testing"

type testObj struct {
	value       int
	resetCalled bool
}

func (t *testObj) Reset() {
	t.value = 0
	t.resetCalled = true
}

func TestPool_NewAndGet(t *testing.T) {
	called := 0

	pool := New(func() *testObj {
		called++
		return &testObj{value: 42}
	})

	obj := pool.Get()
	if obj == nil {
		t.Fatal("expected non-nil object")
	}

	if obj.value != 42 {
		t.Fatalf("unexpected initial value: %d", obj.value)
	}

	if called != 1 {
		t.Fatalf("newFunc should be called once, got %d", called)
	}
}

func TestPool_PutCallsReset(t *testing.T) {
	pool := New(func() *testObj {
		return &testObj{value: 100}
	})

	obj := pool.Get()
	obj.value = 999

	pool.Put(obj)

	if !obj.resetCalled {
		t.Fatal("expected Reset() to be called on Put")
	}

	if obj.value != 0 {
		t.Fatalf("expected value to be reset to 0, got %d", obj.value)
	}
}

func TestPool_ReusesObject(t *testing.T) {
	pool := New(func() *testObj {
		return &testObj{value: 1}
	})

	obj1 := pool.Get()
	obj1.value = 123
	pool.Put(obj1)

	obj2 := pool.Get()

	if obj1 != obj2 {
		t.Fatal("expected object to be reused from pool")
	}

	if !obj2.resetCalled {
		t.Fatal("expected Reset() to be called before reuse")
	}
}

func TestPool_MultipleGets(t *testing.T) {
	called := 0

	pool := New(func() *testObj {
		called++
		return &testObj{}
	})

	obj1 := pool.Get()
	obj2 := pool.Get()

	if obj1 == obj2 {
		t.Fatal("expected different objects for concurrent Gets")
	}

	if called != 2 {
		t.Fatalf("expected newFunc to be called twice, got %d", called)
	}
}
