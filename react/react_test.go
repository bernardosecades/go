package react

import "testing"

// Define a function New() Reactor and the stuff that follows from
// implementing Reactor.
//
// Also define an exported TestVersion with a value that matches
// the internal testVersion here.

const testVersion = 2

// Retired:
//  1 afa5c1278857457403a30479663d26d4e1c8c496

// This is a compile time check to see if you've properly implemented New().
var _ Reactor = New()

// If this test fails and you've proprly defined TestVersion the requirements
// of the tests have changed since you wrote your submission.
func TestTestVersion(t *testing.T) {
	if TestVersion != testVersion {
		t.Fatalf("Found TestVersion = %v, want %v", TestVersion, testVersion)
	}
}

func assertCellValue(t *testing.T, c Cell, expected int, explanation string) {
	observed := c.Value()
	if observed != expected {
		t.Fatalf("%s: expected %d, got %d", explanation, expected, observed)
	}
}

// Setting the value of an input cell changes the observable Value()
func TestSetInput(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	assertCellValue(t, i, 1, "i.Value() doesn't match initial value")
	i.SetValue(2)
	assertCellValue(t, i, 2, "i.Value() doesn't match changed value")
}

// The value of a compute 1 cell is determined by the value of the dependencies.
func TestBasicCompute1(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c := r.CreateCompute1(i, func(v int) int { return v + 1 })
	assertCellValue(t, c, 2, "c.Value() isn't properly computed based on initial input cell value")
	i.SetValue(2)
	assertCellValue(t, c, 3, "c.Value() isn't properly computed based on changed input cell value")
}

// The value of a compute 2 cell is determined by the value of the dependencies.
func TestBasicCompute2(t *testing.T) {
	r := New()
	i1 := r.CreateInput(1)
	i2 := r.CreateInput(2)
	c := r.CreateCompute2(i1, i2, func(v1, v2 int) int { return v1 | v2 })
	assertCellValue(t, c, 3, "c.Value() isn't properly computed based on initial input cell values")
	i1.SetValue(4)
	assertCellValue(t, c, 6, "c.Value() isn't properly computed when first input cell value changes")
	i2.SetValue(8)
	assertCellValue(t, c, 12, "c.Value() isn't properly computed when second input cell value changes")
}

// Compute 2 cells can depend on compute 1 cells.
func TestCompute2Diamond(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c1 := r.CreateCompute1(i, func(v int) int { return v + 1 })
	c2 := r.CreateCompute1(i, func(v int) int { return v - 1 })
	c3 := r.CreateCompute2(c1, c2, func(v1, v2 int) int { return v1 * v2 })
	assertCellValue(t, c3, 0, "c3.Value() isn't properly computed based on initial input cell value")
	i.SetValue(3)
	assertCellValue(t, c3, 8, "c3.Value() isn't properly computed based on changed input cell value")
}

// Compute 1 cells can depend on other compute 1 cells.
func TestCompute1Chain(t *testing.T) {
	r := New()
	inp := r.CreateInput(1)
	var c Cell = inp
	for i := 2; i <= 8; i++ {
		// must save current value of loop variable i for correct behavior.
		// compute function has to use digitToAdd not i.
		digitToAdd := i
		c = r.CreateCompute1(c, func(v int) int { return v*10 + digitToAdd })
	}
	assertCellValue(t, c, 12345678, "c.Value() isn't properly computed based on initial input cell value")
	inp.SetValue(9)
	assertCellValue(t, c, 92345678, "c.Value() isn't properly computed based on changed input cell value")
}

// Compute 2 cells can depend on other compute 2 cells.
func TestCompute2Tree(t *testing.T) {
	r := New()
	ins := make([]InputCell, 3)
	for i, v := range []int{1, 10, 100} {
		ins[i] = r.CreateInput(v)
	}

	add := func(v1, v2 int) int { return v1 + v2 }

	firstLevel := make([]ComputeCell, 2)
	for i := 0; i < 2; i++ {
		firstLevel[i] = r.CreateCompute2(ins[i], ins[i+1], add)
	}

	output := r.CreateCompute2(firstLevel[0], firstLevel[1], add)
	assertCellValue(t, output, 121, "output.Value() isn't properly computed based on initial input cell values")

	for i := 0; i < 3; i++ {
		ins[i].SetValue(ins[i].Value() * 2)
	}

	assertCellValue(t, output, 242, "output.Value() isn't properly computed based on changed input cell values")
}

// Compute cells can have callbacks.
func TestBasicCallback(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c := r.CreateCompute1(i, func(v int) int { return v + 1 })
	var observed []int
	c.AddCallback(func(v int) {
		observed = append(observed, v)
	})
	if len(observed) != 0 {
		t.Fatalf("callback called before changes were made")
	}
	i.SetValue(2)
	if len(observed) != 1 {
		t.Fatalf("callback not called when changes were made")
	}
	if observed[0] != 3 {
		t.Fatalf("callback not called with proper value")
	}
}

// Callbacks and only trigger on change.
func TestOnlyCallOnChanges(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c := r.CreateCompute1(i, func(v int) int {
		if v > 3 {
			return v + 1
		} else {
			return 2
		}
	})
	var observedCalled int
	c.AddCallback(func(int) {
		observedCalled++
	})
	i.SetValue(1)
	if observedCalled != 0 {
		t.Fatalf("observe function called even though input didn't change")
	}
	i.SetValue(2)
	if observedCalled != 0 {
		t.Fatalf("observe function called even though computed value didn't change")
	}
}

// Callbacks can be added and removed.
func TestCallbackAddRemove(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c := r.CreateCompute1(i, func(v int) int { return v + 1 })
	var observed1 []int
	cb1 := c.AddCallback(func(v int) {
		observed1 = append(observed1, v)
	})
	var observed2 []int
	c.AddCallback(func(v int) {
		observed2 = append(observed2, v)
	})
	i.SetValue(2)
	if len(observed1) != 1 || observed1[0] != 3 {
		t.Fatalf("observed1 not properly called")
	}
	if len(observed2) != 1 || observed2[0] != 3 {
		t.Fatalf("observed2 not properly called")
	}
	c.RemoveCallback(cb1)
	i.SetValue(3)
	if len(observed1) != 1 {
		t.Fatalf("observed1 called after removal")
	}
	if len(observed2) != 2 || observed2[1] != 4 {
		t.Fatalf("observed2 not properly called after first callback removal")
	}
}

// Callbacks should only be called once even though
// multiple dependencies have changed.
func TestOnlyCallOnceOnMultipleDepChanges(t *testing.T) {
	r := New()
	i := r.CreateInput(1)
	c1 := r.CreateCompute1(i, func(v int) int { return v + 1 })
	c2 := r.CreateCompute1(i, func(v int) int { return v - 1 })
	c3 := r.CreateCompute1(c2, func(v int) int { return v - 1 })
	c4 := r.CreateCompute2(c1, c3, func(v1, v3 int) int { return v1 * v3 })
	changed4 := 0
	c4.AddCallback(func(int) { changed4++ })
	i.SetValue(3)
	if changed4 < 1 {
		t.Fatalf("callback function was not called")
	} else if changed4 > 1 {
		t.Fatalf("callback function was called too often")
	}
}