package graph

import (
	"github.com/rshelekhov/lazymake/internal/makefile"
)

// Node represents a single target in the dependency graph
type Node struct {
	Target       makefile.Target
	Dependencies []*Node // Outgoing edges: targets this depends on (must run before this)
	Dependents   []*Node // Incoming edges: targets that depend on this (run after this)

	// Graph analysis results (calculated by algorithms)
	Order       int  // Execution order number from topological sort (1, 2, 3...)
	IsCritical  bool // Is this node on the critical path? (longest chain)
	CanParallel bool // Can this run in parallel with its siblings?
}

// Graph represents the complete dependency graph
type Graph struct {
	Nodes map[string]*Node // Map of target name -> Node (for O(1) lookup)
	Roots []*Node          // Entry points: targets that nothing depends on

	// Cycle detection results
	HasCycle   bool
	CycleNodes []string // If there's a cycle, this shows the path

	// Missing dependencies tracking
	MissingDeps map[string][]string // Map of target -> list of missing deps
}

// BuildGraph constructs a dependency graph from parsed Makefile targets
//
// Steps:
// 1. Create a Node for each target
// 2. Wire up the dependency relationships (edges)
// 3. Detect cycles (circular dependencies)
// 4. Calculate execution order (topological sort)
// 5. Identify critical path (longest chain)
// 6. Mark parallel opportunities
// 7. Find root nodes (targets with no dependents)
func BuildGraph(targets []makefile.Target) *Graph {
	g := &Graph{
		Nodes:       make(map[string]*Node),
		MissingDeps: make(map[string][]string),
	}

	// Phase 1: Create all nodes
	for _, target := range targets {
		g.Nodes[target.Name] = &Node{
			Target:       target,
			Dependencies: make([]*Node, 0),
			Dependents:   make([]*Node, 0),
		}
	}

	// Phase 2: Wire up dependencies
	// For each target's dependency list, we find the corresponding Node and link them.
	for _, node := range g.Nodes {
		for _, depName := range node.Target.Dependencies {
			// Try to find the dependency in our graph
			if depNode, exists := g.Nodes[depName]; exists {
				// 1. This node depends on depNode (outgoing edge)
				node.Dependencies = append(node.Dependencies, depNode)

				// 2. depNode is depended upon by this node (incoming edge)
				depNode.Dependents = append(depNode.Dependents, node)
			} else {
				// Dependency not found. Track it for debugging/display purposes
				g.MissingDeps[node.Target.Name] = append(g.MissingDeps[node.Target.Name], depName)

				// Create a placeholder node for visualization
				// If someone's parent isn't in our tree, we still want to
				// show that they exist, even if we don't have details about them
				placeholder := &Node{
					Target: makefile.Target{
						Name:        depName,
						Description: "(external or file dependency)",
					},
					Dependencies: make([]*Node, 0),
					Dependents:   make([]*Node, 0),
				}
				g.Nodes[depName] = placeholder
				node.Dependencies = append(node.Dependencies, placeholder)
			}
		}
	}

	// Phase 3: Check if there are any circular dependencies (A→B→C→A)
	// This would cause infinite loops, so we need to detect and warn about them
	g.HasCycle, g.CycleNodes = detectCycles(g)

	// Only proceed with analysis if there are no cycles
	if !g.HasCycle {
		// Phase 4: Calculate execution order (topological sort)
		calculateExecutionOrder(g)

		// Phase 5: Identify critical path
		identifyCriticalPath(g)

		// Phase 6: Mark parallel opportunities
		identifyParallelOpportunities(g)
	}

	// Phase 7: Find root nodes
	for _, node := range g.Nodes {
		if len(node.Dependents) == 0 {
			g.Roots = append(g.Roots, node)
		}
	}

	return g
}

// detectCycles uses DFS with color-based tracking to detect cycles
func detectCycles(g *Graph) (bool, []string) {
	// Color map: 0 = white (unvisited), 1 = gray (visiting), 2 = black (visited)
	color := make(map[string]int)

	// Parent map: tracks how we got to each node (for reconstructing the cycle path)
	parent := make(map[string]string)

	// These will store the cycle endpoints when we find one
	var cycleStart, cycleEnd string

	// DFS recursive function
	// Returns true if a cycle is found starting from this node
	var dfs func(nodeName string) bool
	dfs = func(nodeName string) bool {
		// Mark this node as GRAY (currently exploring)
		color[nodeName] = 1

		node := g.Nodes[nodeName]

		// Explore all dependencies (outgoing edges)
		for _, dep := range node.Dependencies {
			depName := dep.Target.Name

			// Case 1: Found a GRAY node - this is a back edge, cycle detected
			if color[depName] == 1 {
				cycleStart = depName
				cycleEnd = nodeName
				return true
			}

			// Case 2: Found a WHITE node - need to explore it
			if color[depName] == 0 {
				// Remember that we reached depName from nodeName
				parent[depName] = nodeName

				// Recursively explore this dependency
				if dfs(depName) {
					return true // Cycle found deeper in the tree
				}
			}

			// Case 3: BLACK node - already fully explored, safe to ignore
		}

		// Done exploring this node - mark it BLACK
		color[nodeName] = 2
		return false
	}

	// Try DFS from each unvisited node
	for nodeName := range g.Nodes {
		if color[nodeName] == 0 { // WHITE = unvisited
			if dfs(nodeName) {
				// Cycle found, now reconstruct the path
				cycle := reconstructCyclePath(cycleStart, cycleEnd, parent)
				return true, cycle
			}
		}
	}

	// No cycles found
	return false, nil
}

// reconstructCyclePath builds the cycle path from start to end using parent pointers
//
// We found a cycle from cycleEnd back to cycleStart. The parent map
// tells us how we got to each node. We walk backwards from cycleEnd to cycleStart
// to build the full cycle path.
//
// Example:
//
//	cycleStart = "A", cycleEnd = "C"
//	parent = {B: A, C: B}
//	Result: [A, B, C, A] (showing the complete cycle)
func reconstructCyclePath(cycleStart, cycleEnd string, parent map[string]string) []string {
	// Build path from cycleEnd back to cycleStart
	var path []string
	for curr := cycleEnd; curr != cycleStart; curr = parent[curr] {
		path = append(path, curr)
	}

	// Now reverse the path and add cycleStart at both ends
	// Example: if path is [C, B], we want [A, B, C, A]
	cycle := []string{cycleStart}

	// Add path in reverse order
	for i := len(path) - 1; i >= 0; i-- {
		cycle = append(cycle, path[i])
	}

	// Close the loop by adding cycleStart at the end
	cycle = append(cycle, cycleStart)

	return cycle
}

// calculateExecutionOrder performs topological sort using Kahn's algorithm
//
// Dependencies must run BEFORE their dependents.
//
// Example:
//
//	all → build → deps
//	Order: deps (1), build (2), all (3)
func calculateExecutionOrder(g *Graph) {
	// Step 1: Count in-degrees (number of dependencies for each node)
	inDegree := make(map[string]int)
	for name, node := range g.Nodes {
		inDegree[name] = len(node.Dependencies)
	}

	// Step 2: Find all nodes with in-degree 0 (no dependencies)
	queue := make([]*Node, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, g.Nodes[name])
		}
	}

	// Step 3: Process nodes level by level (BFS)
	order := 0
	for len(queue) > 0 {
		// Increment order ONCE per level (all nodes at same level can run in parallel)
		order++

		// Get all nodes at current level
		levelSize := len(queue)

		// Process all nodes at this level - they all get the SAME order number
		for range levelSize {
			node := queue[0]
			queue = queue[1:]

			// Assign the same order to all nodes at this level
			node.Order = order

			// Step 4: "Remove" this node by updating dependent nodes
			for _, dependent := range node.Dependents {
				inDegree[dependent.Target.Name]--

				if inDegree[dependent.Target.Name] == 0 {
					queue = append(queue, dependent)
				}
			}
		}
	}
}

// identifyCriticalPath finds the longest path through the graph
//
// IMPORTANT: Only marks targets as critical if they're part of an actual dependency chain.
// Standalone targets (no deps, no dependents) are NOT marked as critical since the
// concept of "critical path" only makes sense when there are dependencies to coordinate.
//
// Example:
//
//	Path 1: all → build → compile → deps  (length 4)
//	Path 2: all → test → compile → deps   (length 4)
//	Both are critical paths (tied for longest)
//
// But:
//	clean (standalone, no deps) → NOT critical (it's just a simple target)
func identifyCriticalPath(g *Graph) {
	depth := make(map[string]int)

	var calculateDepth func(node *Node) int
	calculateDepth = func(node *Node) int {
		nodeName := node.Target.Name

		if d, exists := depth[nodeName]; exists {
			return d
		}

		if len(node.Dependencies) == 0 {
			depth[nodeName] = 0
			return 0
		}

		// Recursive case: depth = max(child depths) + 1
		maxDepth := 0
		for _, dep := range node.Dependencies {
			depDepth := calculateDepth(dep)
			if depDepth > maxDepth {
				maxDepth = depDepth
			}
		}

		depth[nodeName] = maxDepth + 1
		return maxDepth + 1
	}

	maxDepth := 0
	for _, node := range g.Nodes {
		d := calculateDepth(node)
		if d > maxDepth {
			maxDepth = d
		}
	}

	// Only mark critical path if there's an actual dependency chain
	// If maxDepth == 0, all targets are independent - no critical path exists
	if maxDepth == 0 {
		return
	}

	// Mark nodes on the critical path
	// Start from nodes with max depth and work backwards
	for _, node := range g.Nodes {
		// Only mark as critical if:
		// 1. It has maximum depth (is at the top of the chain)
		// 2. It has dependencies (part of a chain, not standalone)
		if depth[node.Target.Name] == maxDepth && len(node.Dependencies) > 0 {
			node.IsCritical = true
			markCriticalPathDown(node, depth)
		}
	}
}

// markCriticalPathDown marks nodes along the critical path from a critical node down to leaves
func markCriticalPathDown(node *Node, depth map[string]int) {
	if len(node.Dependencies) == 0 {
		return
	}

	nodeName := node.Target.Name
	nodeDepth := depth[nodeName]

	// Find dependencies that are on the critical path
	// They must have depth = current depth - 1
	for _, dep := range node.Dependencies {
		depName := dep.Target.Name
		if depth[depName] == nodeDepth-1 {
			// This dependency is on the critical path
			dep.IsCritical = true
			// Recursively mark the path down from this dependency
			markCriticalPathDown(dep, depth)
		}
	}
}

// identifyParallelOpportunities marks targets that can run in parallel
//
// IMPORTANT: Only marks targets as parallel if they're part of actual build chains.
// Standalone targets (no dependencies) are NOT marked as parallel since the concept
// of "parallelization" only makes sense when there are dependencies to coordinate.
//
// Targets can run in parallel if they:
// 1. Have the same execution order (from topological sort)
// 2. Have at least one dependency (part of a build chain)
// 3. Don't depend on each other
//
// Example:
//
//	all → build → deps
//	  ↓
//	all → test → deps
//
// "build" and "test" both have order 2 and have deps, so they can run in parallel
//
// But:
//	clean, lint, format (all independent, no deps) → NOT parallel (just standalone)
//
// This tells users where they can speed things up with parallel execution
func identifyParallelOpportunities(g *Graph) {
	// Group nodes by execution order
	// All nodes with the same order number are at the same "level"
	// and can potentially run in parallel
	orderGroups := make(map[int][]*Node)
	for _, node := range g.Nodes {
		// Only consider nodes that have dependencies (part of a build chain)
		if len(node.Dependencies) > 0 {
			order := node.Order
			orderGroups[order] = append(orderGroups[order], node)
		}
	}

	// Mark parallel opportunities
	// If there's more than one node at a given level, they can all run in parallel
	for _, group := range orderGroups {
		if len(group) > 1 {
			for _, node := range group {
				node.CanParallel = true
			}
		}
	}
}

// GetSubgraph extracts a portion of the graph centered on a specific target
//
// This is useful for viewing just one target's dependencies without showing
// the entire graph. You can also limit the depth to avoid showing too much.
//
// Parameters:
//   - targetName: The root target to center the subgraph on
//   - maxDepth: How many levels deep to go (-1 = unlimited)
//   - 0 = just the target itself
//   - 1 = target + direct dependencies
//   - 2 = target + dependencies + their dependencies
//   - etc.
//
// Algorithm: BFS (Breadth-First Search) with depth tracking
//
// Example:
//
//	all → build → compile → deps
//	  ↓     ↓
//	test → lint
//
// GetSubgraph("build", 1) returns:
//
//	build → compile
//	  ↓
//	lint
func (g *Graph) GetSubgraph(targetName string, maxDepth int) *Graph {
	// Handle unlimited depth case
	if maxDepth < 0 {
		maxDepth = 999999
	}

	// Check if the target exists
	rootNode, exists := g.Nodes[targetName]
	if !exists {
		// Return empty graph if target not found
		return &Graph{
			Nodes:       make(map[string]*Node),
			MissingDeps: make(map[string][]string),
		}
	}

	// Create new subgraph with same cycle info
	subgraph := &Graph{
		Nodes:       make(map[string]*Node),
		Roots:       []*Node{rootNode},
		HasCycle:    g.HasCycle,
		CycleNodes:  g.CycleNodes,
		MissingDeps: make(map[string][]string),
	}

	// BFS queue item: tracks node and its depth from root
	type queueItem struct {
		node  *Node
		depth int
	}

	// Initialize BFS
	queue := []queueItem{{rootNode, 0}}
	visited := make(map[string]bool)

	// BFS traversal
	for len(queue) > 0 {
		// Dequeue
		item := queue[0]
		queue = queue[1:]

		nodeName := item.node.Target.Name

		// Skip if already visited (prevents duplicates)
		if visited[nodeName] {
			continue
		}

		// Mark as visited
		visited[nodeName] = true

		// Add node to subgraph
		subgraph.Nodes[nodeName] = item.node

		// Add dependencies to queue if we haven't reached max depth
		if item.depth < maxDepth {
			for _, dep := range item.node.Dependencies {
				if !visited[dep.Target.Name] {
					queue = append(queue, queueItem{dep, item.depth + 1})
				}
			}
		}
	}

	return subgraph
}
