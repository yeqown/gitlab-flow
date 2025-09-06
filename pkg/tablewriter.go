package pkg

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func NewTableWriter() *tablewriter.Table {
	cnf := tablewriter.Config{
		Header: tw.CellConfig{
			Formatting: tw.CellFormatting{
				AutoFormat: tw.On,
			},
			Alignment: tw.CellAlignment{
				Global: tw.AlignCenter,
			},
		},
		Row: tw.CellConfig{
			Alignment: tw.CellAlignment{
				Global: tw.AlignLeft,
			},
			Formatting: tw.CellFormatting{
				MergeMode: tw.MergeBoth,
			},
		},
		// Debug: false,
	}

	style := tw.NewSymbolCustom("Tech").
		WithRow("-").
		WithColumn("|").
		WithTopLeft("+").
		WithTopMid("##").
		WithTopRight("+").
		WithMidLeft("+").
		WithCenter("+").
		WithMidRight("+").
		WithBottomLeft("+").
		WithBottomMid("##").
		WithBottomRight("+")

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: style})),
		tablewriter.WithConfig(cnf),
	)

	return table
}
