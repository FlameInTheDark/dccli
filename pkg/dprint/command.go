package dprint

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	CommandDescribeTemplate = `
ID: {{.ID}}
Name: {{.Name}}

`
	GuildDescribeTemplate = `
ID: {{.ID}}
Name: {{.Name}}
Owner ID: {{.OwnerID}}
Owner Name: {{.OwnerName}}
Owner Nickname: {{.OwnerNickname}}
Emojis Count: {{.EmojisCount}}
Roles Count: {{.RolesCount}}
`
)

func PrintTemplate(tpl string, v interface{}) error {
	newtpl, err := template.New("print").Parse(tpl)
	if err != nil {
		return err
	}
	var raw []byte
	buff := bytes.NewBuffer(raw)
	err = newtpl.Execute(buff, v)
	if err != nil {
		return err
	}

	fmt.Println(buff.String())
	return nil
}
