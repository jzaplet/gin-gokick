package literals

import "text/template/parse"

func Of(node parse.Node, fn string) []string {
	var literals []string

	for _, call := range Calls(node, fn) {
		for _, arg := range call.Args[1:] {
			if text, ok := arg.(*parse.StringNode); ok {
				literals = append(literals, text.Text)
			}
		}
	}

	return literals
}

func Calls(node parse.Node, fn string) []*parse.CommandNode {
	return appendNode(nil, node, fn)
}

func appendNode(calls []*parse.CommandNode, node parse.Node, fn string) []*parse.CommandNode {
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return calls
		}

		for _, child := range n.Nodes {
			calls = appendNode(calls, child, fn)
		}
	case *parse.ActionNode:
		calls = appendNode(calls, n.Pipe, fn)
	case *parse.TemplateNode:
		calls = appendNode(calls, n.Pipe, fn)
	case *parse.IfNode:
		calls = appendBranch(calls, &n.BranchNode, fn)
	case *parse.RangeNode:
		calls = appendBranch(calls, &n.BranchNode, fn)
	case *parse.WithNode:
		calls = appendBranch(calls, &n.BranchNode, fn)
	case *parse.PipeNode:
		if n == nil {
			return calls
		}

		for _, cmd := range n.Cmds {
			calls = appendNode(calls, cmd, fn)
		}
	case *parse.CommandNode:
		calls = appendCommand(calls, n, fn)
	}

	return calls
}

func appendCommand(calls []*parse.CommandNode, cmd *parse.CommandNode, fn string) []*parse.CommandNode {
	if isFunc(cmd.Args[0], fn) {
		calls = append(calls, cmd)
	}

	for _, arg := range cmd.Args[1:] {
		if isFunc(arg, fn) {
			calls = append(calls, &parse.CommandNode{
				NodeType: parse.NodeCommand,

				Pos: arg.Position(),

				Args: []parse.Node{arg},
			})
		}
	}

	for _, arg := range cmd.Args {
		calls = appendNode(calls, arg, fn)
	}

	return calls
}

func isFunc(node parse.Node, fn string) bool {
	ident, ok := node.(*parse.IdentifierNode)

	return ok && ident.Ident == fn
}

func appendBranch(calls []*parse.CommandNode, branch *parse.BranchNode, fn string) []*parse.CommandNode {
	calls = appendNode(calls, branch.Pipe, fn)
	calls = appendNode(calls, branch.List, fn)

	return appendNode(calls, branch.ElseList, fn)
}
