package grider

import (
	"testing"
)

func Test_cut(t *testing.T) {
	cases := []struct {
		bf       BaseFilter
		slicelen int
		from     int
		to       int
	}{
		{BaseFilter{PageNumber: 0, PageSize: 10}, 1, 0, 1},
		{BaseFilter{PageNumber: 1, PageSize: 10}, 1, 0, 0},
	}

	for i := range cases {
		from, to := cases[i].bf.cut(cases[i].slicelen)
		if from != cases[i].from || to != cases[i].to {
			t.Errorf("case failed: %d. expected: (%d,%d) got: (%d,%d)", i, cases[i].from, cases[i].to, from, to)
		}
	}
}
