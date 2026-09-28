package main

import (
	"fmt"
	"os"

	"go.i3wm.org/i3/v4"
)

type Edge string

const (
	LEFT  Edge = "LEFT"
	RIGHT Edge = "RIGHT"
)

func getSpelledName(wsName string) string {
	spellings := map[string]string{
		"1": "One", "2": "Two", "3": "Three", "4": "Four", "5": "Five",
		"6": "Six", "7": "Seven", "8": "Eight", "9": "Nine", "10": "Ten",
	}

	// If it's a standard number, return the spelling
	if spelled, exists := spellings[wsName]; exists {
		return spelled
	}

	// Fallback just in case you use named workspaces like "Web"
	return wsName
}

// findNode recursively searches for a node matching the predicate
func findNode(node *i3.Node, predicate func(*i3.Node) bool) *i3.Node {
	if predicate(node) {
		return node
	}
	for i := range node.Nodes {
		if res := findNode(node.Nodes[i], predicate); res != nil {
			return res
		}
	}
	return nil
}

func getFocusedWorkspaceName() string {
	workspaces, err := i3.GetWorkspaces()
	if err == nil {
		for _, w := range workspaces {
			if w.Focused {
				return w.Name
			}
		}
	}
	return ""
}

func getEdgeNode(workspace string, edge Edge) *i3.Node {
	tree, _ := i3.GetTree()
	wNode := findNode(tree.Root, func(n *i3.Node) bool { return n.Type == i3.WorkspaceNode && n.Name == workspace })

	if wNode == nil || len(wNode.Nodes) == 0 {
		return nil
	} else {

		var wContainers = wNode.Nodes
		if edge == LEFT {
			return wContainers[0]
		} else {
			return wContainers[len(wContainers)-1]
		}
	}
}

func moveContainerToWorkspace(workspace string, container *i3.Node, edge Edge) {
	tree, _ := i3.GetTree()

	wNode := tree.Root.FindChild(func(n *i3.Node) bool { return n.Type == i3.WorkspaceNode && n.Name == workspace })
	if wNode == nil || len(wNode.Nodes) == 0 {
		i3.RunCommand(fmt.Sprintf("[con_id=\"%d\"] move container to workspace %s", container.ID, workspace))
		return
	} else {

		markContainer := wNode.Nodes[len(wNode.Nodes)-1]
		if edge == LEFT {
			markContainer = wNode.Nodes[0]
		}

		fmt.Println(markContainer.Layout)

		cmd := fmt.Sprintf(`[con_id="%d"] mark "_temp_target"; `+
			`[con_id="%d"] move container to mark "_temp_target"; `+
			`[con_id="%d"] unmark "_temp_target";`, markContainer.ID, container.ID, markContainer.ID)

		if edge == LEFT {
			cmd += fmt.Sprintf(` [con_id="%d"] move left;`, container.ID)
		}
		if markContainer.Name == "" {

			if edge == LEFT {
				cmd += fmt.Sprintf(` [con_id="%d"] move left;`, container.ID)
			} else {
				cmd += fmt.Sprintf(` [con_id="%d"] move right`, container.ID)
			}
		}
		i3.RunCommand(cmd)
	}
}

func Scroll(direction Edge) {

	activeWsName := getFocusedWorkspaceName()
	if activeWsName == "" {
		return
	}

	lwName := "L" + getSpelledName(activeWsName)
	rwName := "R" + getSpelledName(activeWsName)

	var oldNode *i3.Node

	var targetWS, sourceWS string

	if direction == LEFT {
		targetWS = rwName
		sourceWS = lwName
		oldNode = getEdgeNode(activeWsName, RIGHT)
	} else {
		targetWS = lwName
		sourceWS = rwName
		oldNode = getEdgeNode(activeWsName, LEFT)
	}

	if oldNode == nil {
		return
	}

	newNode := getEdgeNode(sourceWS, RIGHT)
	if newNode == nil {
		return
	}

	if direction == LEFT {
		moveContainerToWorkspace(targetWS, oldNode, RIGHT)
		moveContainerToWorkspace(activeWsName, newNode, LEFT)
	} else {
		moveContainerToWorkspace(targetWS, oldNode, RIGHT)
		moveContainerToWorkspace(activeWsName, newNode, RIGHT)
	}

	i3.RunCommand(fmt.Sprintf("[con_id=\"%d\"] focus", newNode.ID))

}

func Move(direction Edge) {

	activeWsName := getFocusedWorkspaceName()
	if activeWsName == "" {
		return
	}

	var targetWs string
	var oldNode = getEdgeNode(activeWsName, direction)

	if oldNode == nil {
		return
	}
	if direction == RIGHT {
		targetWs = "R" + getSpelledName(activeWsName)
	} else {
		targetWs = "L" + getSpelledName(activeWsName)
	}

	moveContainerToWorkspace(targetWs, oldNode, RIGHT)
}

func Collect(direction Edge) {
	activeWsName := getFocusedWorkspaceName()
	if activeWsName == "" {
		return
	}

	lwName := "L" + getSpelledName(activeWsName)
	rwName := "R" + getSpelledName(activeWsName)

	var sourceWS string

	if direction == LEFT {
		sourceWS = lwName
	} else {
		sourceWS = rwName
	}

	newNode := getEdgeNode(sourceWS, RIGHT)
	if newNode == nil {
		return
	}

	if direction == LEFT {
		moveContainerToWorkspace(activeWsName, newNode, LEFT)
	} else {
		moveContainerToWorkspace(activeWsName, newNode, RIGHT)
	}

}

func AutoSplit(direction Edge) {
	tree, _ := i3.GetTree()
	focusedNode := tree.Root.FindChild(func(n *i3.Node) bool { return n.Focused })
	if focusedNode == nil {
		return
	}

	if direction == LEFT {
		i3.RunCommand("focus left")
	} else {
		i3.RunCommand("focus right")
	}

	tree, _ = i3.GetTree()
	newFocusedNode := tree.Root.FindChild(func(n *i3.Node) bool { return n.Focused })

	if newFocusedNode == nil || newFocusedNode.ID == focusedNode.ID {
		return
	}

	cmd := fmt.Sprintf(`[con_id="%d"] mark "_auto_split_target"; [con_id="%d"] split v; `, newFocusedNode.ID, newFocusedNode.ID)
	cmd += fmt.Sprintf(`[con_id="%d"] move container to mark "_auto_split_target"; `, focusedNode.ID)

	if direction == RIGHT {
		cmd += fmt.Sprintf(`[con_id="%d"] move up; `, focusedNode.ID)
	}

	cmd += fmt.Sprintf(`[con_id="%d"] unmark "_auto_split_target"; `, newFocusedNode.ID)
	cmd += fmt.Sprintf(`[con_id="%d"] focus`, focusedNode.ID)

	i3.RunCommand(cmd)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: i3-scroller <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scroll-left":
		Scroll(LEFT)
	case "scroll-right":
		Scroll(RIGHT)
	case "move-left":
		Move(LEFT)
	case "move-right":
		Move(RIGHT)
	case "collect-left":
		Collect(LEFT)
	case "collect-right":
		Collect(RIGHT)
	case "auto-split-left":
		AutoSplit(LEFT)
	case "auto-split-right":
		AutoSplit(RIGHT)
	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
