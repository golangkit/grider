package grid

import (
	"encoding/json"
	"fmt"
	"testing"

	"bitbucket.org/rtg365/lms/backend/service/meter"
)

func TestConvert(t *testing.T) {

	s := []MeterGridItem{
		{ID: 1, SerialNumber: "60890011", SpecName: "Incotex 205", m: &meter.Meter{SerialNumber: "XXX001"}},
	}

	grid := ConvertSliceOfStructToGrid("inventory.meters.", s)
	for i := range grid.Columns {
		fmt.Printf("%#v\n", grid.Columns[i])
	}

	buf, _ := json.Marshal(grid)
	fmt.Println(string(buf))
}

func TestMetersConvert(t *testing.T) {

}
