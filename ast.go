package tpl

import (
	"text/template/parse"
)

func extractTemplateField(name string, tree *parse.Tree) []string {
	var fields []string
	for _, n := range tree.Root.Nodes {
		fields = append(fields, extractFields(n)...)
	}

	return fields
}

func extractFields(node parse.Node) []string {
	var fields []string
	switch v := node.(type) {
	case *parse.ActionNode:
		for _, cmd := range v.Pipe.Cmds {
			for idx, arg := range cmd.Args {
				if arg.Type() == parse.NodeField {
					fields = append(fields, arg.String())
				} else if arg.Type() == parse.NodeIdentifier && arg.String() == "tpltype" {
					types := "@type:"
					for i := idx + 1; i < len(cmd.Args); i++ {
						if cmd.Args[i].Type() == parse.NodeString {
							types += cmd.Args[i].String() + ","
						}
					}

					fields = append(fields, types)
				}
			}
		}
	}

	return fields
}
