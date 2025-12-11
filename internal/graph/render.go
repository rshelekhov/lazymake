package graph

import (
	"fmt"
	"strings"

	"github.com/rshelekhov/lazymake/internal/util"
)

// TreeRenderer controls what annotations to show in the tree
type TreeRenderer struct {
	ShowOrder    bool // Show execution order numbers [1] [2] [3]
	ShowCritical bool // Show critical path marker ★
	ShowParallel bool // Show parallel marker ∥
}

// RenderTree returns a string representation of the graph as an ASCII tree
//
// It's converting our graph into a visual tree using box-drawing characters:
//
//	├── (branch, not last)
//	└── (branch, last)
//	│   (vertical continuation)
//
// Example output:
//
//	all [3] ★
//	├── build [2] ∥
//	│   └── deps [1]
//	└── test [2] ∥
//	    └── deps [1] (see above)
func (g *Graph) RenderTree(renderer TreeRenderer) string {
	var builder strings.Builder

	// Handle cycles first - if there's a cycle, we can't render a proper tree
	if g.HasCycle {
		util.WriteString(&builder, "⚠️  Circular dependency detected!\n")
		util.WriteString(&builder, "Cycle: ")
		util.WriteString(&builder, strings.Join(g.CycleNodes, " → "))
		util.WriteString(&builder, "\n\n")
		util.WriteString(&builder, "Fix the circular dependency in your Makefile before visualizing the graph.\n")
		return builder.String()
	}

	// Handle empty graph
	if len(g.Nodes) == 0 {
		return "No targets found in Makefile.\n"
	}

	// Render from each root node
	visited := make(map[string]bool)
	for i, root := range g.Roots {
		if i > 0 {
			util.WriteString(&builder, "\n") // Blank line between separate trees
		}
		renderNode(root, "", true, &builder, renderer, visited)
	}

	return builder.String()
}

// renderNode recursively renders a node and its dependencies as an ASCII tree
//
// This is a recursive DFS traversal that builds the tree string as it goes.
// The prefix grows longer as we go deeper, creating the indentation.
func renderNode(
	node *Node, // The node to render
	prefix string, // The string to print before this node (contains │ and spaces for indentation)
	isLast bool, // Is this the last child of its parent? (affects which branch character to use)
	builder *strings.Builder, // Where to write the output
	renderer TreeRenderer, // Controls which annotations to show
	visited map[string]bool, // Tracks nodes we've already rendered (prevents infinite loops)
) {
	nodeName := node.Target.Name

	// Check if we've already rendered this node
	// In graphs with shared dependencies (diamond pattern), we might encounter
	// the same node multiple times. We show it once fully, then just reference it
	// with "(see above)" for subsequent encounters.
	if visited[nodeName] {
		// This node was already rendered - just show a reference
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		util.WriteString(builder, prefix+connector+nodeName+" (see above)\n")
		return
	}

	// Mark as visited
	visited[nodeName] = true

	// Determine which branch character to use
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Build the node display string with all requested annotations
	nodeStr := buildNodeString(node, renderer)

	// Write this node's line
	util.WriteString(builder, prefix+connector+nodeStr+"\n")

	// Prepare prefix for children
	extension := "│   "
	if isLast {
		extension = "    "
	}

	// Recursively render dependencies (children)
	deps := node.Dependencies
	for i, dep := range deps {
		isLastDep := i == len(deps)-1
		renderNode(dep, prefix+extension, isLastDep, builder, renderer, visited)
	}
}

// buildNodeString creates the display string for a node with all annotations
//
// Example outputs:
//
//	"build"                    (just the name)
//	"build [2]"                (with order)
//	"build [2] ★"              (order + critical)
//	"build [2] ★ ∥"            (order + critical + parallel)
//	"build — Build the app"    (with description)
//	"build [2] ★ ∥ — Build"    (everything!)
func buildNodeString(node *Node, renderer TreeRenderer) string {
	var parts []string

	// Start with the target name
	parts = append(parts, node.Target.Name)

	// Add execution order [N]
	if renderer.ShowOrder && node.Order > 0 {
		parts = append(parts, fmt.Sprintf("[%d]", node.Order))
	}

	// Add critical path marker ★
	if renderer.ShowCritical && node.IsCritical {
		parts = append(parts, "★")
	}

	// Add parallel marker ∥
	if renderer.ShowParallel && node.CanParallel {
		parts = append(parts, "∥")
	}

	result := strings.Join(parts, " ")

	// Add description if present
	if node.Target.Description != "" {
		result += " — " + node.Target.Description
	}

	return result
}

// RenderLegend returns a legend explaining the symbols used in the tree
//
// Example: "Legend: [N] = execution order, ★ = critical path, ∥ = can run in parallel"
func RenderLegend(showOrder, showCritical, showParallel bool) string {
	var parts []string

	if showOrder {
		parts = append(parts, "[N] = execution order")
	}
	if showCritical {
		parts = append(parts, "★ = critical path")
	}
	if showParallel {
		parts = append(parts, "∥ = can run in parallel")
	}

	if len(parts) == 0 {
		return ""
	}

	return "Legend: " + strings.Join(parts, ", ")
}
