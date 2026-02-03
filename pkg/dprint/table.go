package dprint

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

func Table(header []string, data [][]string) {
	var raw []byte
	buff := bytes.NewBuffer(raw)
	table := tablewriter.NewWriter(buff)
	table.SetHeader(header)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	fmt.Println(buff.String())
}
