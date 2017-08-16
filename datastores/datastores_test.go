package datastores

import "testing"

func TestIterationsInSeconds(t *testing.T) {
	run1 := IterationsInSeconds(10, 1)
	var expected int64 = 10
	if run1 != expected {
		t.Errorf("IterationsInSeconds calculation is incorrect, expected: %d, actual: %d", run1)
	}

	run2 := IterationsInSeconds(10, .5)
	expected = 5
	if run2 != expected {
		t.Errorf("IterationsInSeconds calculation is incorrect, expected: %d, actual: %d", expected, run1)
	}
}
