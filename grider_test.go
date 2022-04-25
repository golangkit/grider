package grider_test

import (
	"testing"

	"github.com/golangkit/grider"
)

func TestNewPage(t *testing.T) {

	a := grider.ActionSet{
		"Edit": grider.Action{Code: "Edit"},
	}

	g := grider.Grid{
		Columns:    []grider.Column{{Name: "Name"}},
		Rows:       [][]string{{"Robert"}},
		RowActions: [][]grider.ActionCode{{"Edit"}},
	}

	p := grider.Page{
		Widgets: []grider.Widgeter{grider.GridWidget{
			Grid: &g,
		}},
	}

	p.AssignActionSet(a)

	t.Logf("%#v", p)
}
