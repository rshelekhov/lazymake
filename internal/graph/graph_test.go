package graph

import (
	"testing"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

// TestBuildGraph tests basic graph construction
func TestBuildGraph(t *testing.T) {
	// Create a simple dependency chain:
	// all → build → deps
	//   ↓
	// test → build
	//
	// Execution order should be: deps (1), build (2), then test and all can run (3)
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build", "test"}},
		{Name: "build", Dependencies: []string{"deps"}},
		{Name: "test", Dependencies: []string{"build"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// Test 1: All nodes created
	if len(g.Nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(g.Nodes))
	}

	// Test 2: Check dependency wiring (outgoing edges)
	buildNode := g.Nodes["build"]
	if len(buildNode.Dependencies) != 1 {
		t.Errorf("build should have 1 dependency, got %d", len(buildNode.Dependencies))
	}
	if buildNode.Dependencies[0].Target.Name != "deps" {
		t.Errorf("build should depend on deps, got %s", buildNode.Dependencies[0].Target.Name)
	}

	// Test 3: Check dependent wiring (incoming edges)
	depsNode := g.Nodes["deps"]
	if len(depsNode.Dependents) != 1 {
		t.Errorf("deps should have 1 dependent, got %d", len(depsNode.Dependents))
	}
	if depsNode.Dependents[0].Target.Name != "build" {
		t.Errorf("deps should be depended on by build, got %s", depsNode.Dependents[0].Target.Name)
	}

	// Test 4: Check root nodes
	if len(g.Roots) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(g.Roots))
	}
	if g.Roots[0].Target.Name != "all" {
		t.Errorf("Expected 'all' as root, got %s", g.Roots[0].Target.Name)
	}

	// Test 5: No cycles in this graph
	if g.HasCycle {
		t.Error("Graph should not have cycles")
	}
}

// TestBuildGraphMissingDependency tests handling of missing dependencies
func TestBuildGraphMissingDependency(t *testing.T) {
	targets := []makefile.Target{
		{Name: "build", Dependencies: []string{"nonexistent"}},
	}

	g := BuildGraph(targets)

	// Should create placeholder node for missing dependency
	if _, exists := g.Nodes["nonexistent"]; !exists {
		t.Error("Should create placeholder for missing dependency")
	}

	// Should track the missing dependency
	if len(g.MissingDeps) != 1 {
		t.Errorf("Expected 1 missing dependency, got %d", len(g.MissingDeps))
	}

	if missing, ok := g.MissingDeps["build"]; ok {
		if len(missing) != 1 || missing[0] != "nonexistent" {
			t.Errorf("Expected missing dep 'nonexistent', got %v", missing)
		}
	} else {
		t.Error("Missing dependency not tracked")
	}

	// Placeholder should have description indicating it's external
	placeholder := g.Nodes["nonexistent"]
	if placeholder.Target.Description != "(external or file dependency)" {
		t.Errorf("Placeholder should have description, got %q", placeholder.Target.Description)
	}
}

// TestBuildGraphNoDependencies tests graph with independent targets
func TestBuildGraphNoDependencies(t *testing.T) {
	targets := []makefile.Target{
		{Name: "clean", Dependencies: nil},
		{Name: "lint", Dependencies: nil},
		{Name: "format", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// All targets should be roots (nothing depends on them)
	if len(g.Roots) != 3 {
		t.Errorf("Expected 3 root nodes, got %d", len(g.Roots))
	}

	// All targets should have no dependencies
	for _, node := range g.Nodes {
		if len(node.Dependencies) != 0 {
			t.Errorf("Target %s should have no dependencies, got %d",
				node.Target.Name, len(node.Dependencies))
		}
		if len(node.Dependents) != 0 {
			t.Errorf("Target %s should have no dependents, got %d",
				node.Target.Name, len(node.Dependents))
		}
	}
}

// TestBuildGraphDiamond tests a diamond dependency pattern
//
//	  all
//	 /   \
//	A     B
//	 \   /
//	  deps
//
// This is common in real Makefiles and tests that we handle shared dependencies
func TestBuildGraphDiamond(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"deps"}},
		{Name: "B", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// deps should have TWO dependents (both A and B depend on it)
	depsNode := g.Nodes["deps"]
	if len(depsNode.Dependents) != 2 {
		t.Errorf("deps should have 2 dependents, got %d", len(depsNode.Dependents))
	}

	// all should have TWO dependencies
	allNode := g.Nodes["all"]
	if len(allNode.Dependencies) != 2 {
		t.Errorf("all should have 2 dependencies, got %d", len(allNode.Dependencies))
	}

	// Verify A and B each depend on deps
	aNode := g.Nodes["A"]
	if len(aNode.Dependencies) != 1 || aNode.Dependencies[0].Target.Name != "deps" {
		t.Error("A should depend on deps")
	}

	bNode := g.Nodes["B"]
	if len(bNode.Dependencies) != 1 || bNode.Dependencies[0].Target.Name != "deps" {
		t.Error("B should depend on deps")
	}
}

// TestDetectCycleSimple tests detection of a simple 3-node cycle
// CONCEPT: A → B → C → A (simple cycle)
func TestDetectCycleSimple(t *testing.T) {
	targets := []makefile.Target{
		{Name: "A", Dependencies: []string{"B"}},
		{Name: "B", Dependencies: []string{"C"}},
		{Name: "C", Dependencies: []string{"A"}}, // Creates the cycle!
	}

	g := BuildGraph(targets)

	// Should detect the cycle
	if !g.HasCycle {
		t.Error("Expected cycle to be detected")
	}

	// Should identify the nodes in the cycle
	if len(g.CycleNodes) == 0 {
		t.Error("Expected cycle nodes to be identified")
	}

	// The cycle should contain all three nodes
	// (The exact path might vary, but should show the cycle)
	t.Logf("Cycle detected: %v", g.CycleNodes)

	// Verify cycle path starts and ends with the same node
	if len(g.CycleNodes) >= 2 {
		first := g.CycleNodes[0]
		last := g.CycleNodes[len(g.CycleNodes)-1]
		if first != last {
			t.Errorf("Cycle path should start and end with same node, got %s and %s", first, last)
		}
	}
}

// TestDetectCycleSelfReference tests detection of a node depending on itself
// CONCEPT: A → A (simplest possible cycle)
func TestDetectCycleSelfReference(t *testing.T) {
	targets := []makefile.Target{
		{Name: "A", Dependencies: []string{"A"}}, // Depends on itself!
	}

	g := BuildGraph(targets)

	if !g.HasCycle {
		t.Error("Expected self-reference to be detected as a cycle")
	}

	t.Logf("Self-reference cycle: %v", g.CycleNodes)
}

// TestDetectCycleComplex tests detection in a graph with a cycle and non-cycle parts
//
//	clean (OK, no deps)
//	all → build → A → B → C → A (cycle here)
func TestDetectCycleComplex(t *testing.T) {
	targets := []makefile.Target{
		{Name: "clean", Dependencies: nil}, // No cycle here
		{Name: "all", Dependencies: []string{"build"}},
		{Name: "build", Dependencies: []string{"A"}},
		{Name: "A", Dependencies: []string{"B"}},
		{Name: "B", Dependencies: []string{"C"}},
		{Name: "C", Dependencies: []string{"A"}}, // Cycle: A→B→C→A
	}

	g := BuildGraph(targets)

	if !g.HasCycle {
		t.Error("Expected cycle to be detected in complex graph")
	}

	t.Logf("Cycle in complex graph: %v", g.CycleNodes)

	// The cycle nodes should be A, B, C (and A again to close the loop)
	// The exact order might vary, but should include all three
	cycleMap := make(map[string]bool)
	for _, node := range g.CycleNodes {
		cycleMap[node] = true
	}

	expectedInCycle := []string{"A", "B", "C"}
	for _, expected := range expectedInCycle {
		if !cycleMap[expected] {
			t.Errorf("Expected %s to be in cycle nodes", expected)
		}
	}
}

// TestNoCycle verifies that graphs without cycles are correctly identified
func TestNoCycle(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build", "test"}},
		{Name: "build", Dependencies: []string{"compile"}},
		{Name: "test", Dependencies: []string{"compile"}},
		{Name: "compile", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	if g.HasCycle {
		t.Errorf("No cycle should be detected, but got cycle: %v", g.CycleNodes)
	}

	if len(g.CycleNodes) != 0 {
		t.Errorf("CycleNodes should be empty, got %v", g.CycleNodes)
	}
}

// TestDetectCycleTwoNode tests the smallest multi-node cycle
// CONCEPT: A → B → A (two nodes depending on each other)
func TestDetectCycleTwoNode(t *testing.T) {
	targets := []makefile.Target{
		{Name: "A", Dependencies: []string{"B"}},
		{Name: "B", Dependencies: []string{"A"}}, // Mutual dependency!
	}

	g := BuildGraph(targets)

	if !g.HasCycle {
		t.Error("Expected mutual dependency to be detected as cycle")
	}

	t.Logf("Two-node cycle: %v", g.CycleNodes)

	// Should have both A and B in the cycle
	cycleMap := make(map[string]bool)
	for _, node := range g.CycleNodes {
		cycleMap[node] = true
	}

	if !cycleMap["A"] || !cycleMap["B"] {
		t.Error("Expected both A and B in cycle")
	}
}

// TestTopologicalSortSimple tests basic execution order
// CONCEPT: Dependencies should have lower order numbers than their dependents
func TestTopologicalSortSimple(t *testing.T) {
	// Simple chain: all → build → deps
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build"}},
		{Name: "build", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// Verify no cycles (topological sort only works on acyclic graphs)
	if g.HasCycle {
		t.Fatal("Unexpected cycle in graph")
	}

	// Check execution order
	depsOrder := g.Nodes["deps"].Order
	buildOrder := g.Nodes["build"].Order
	allOrder := g.Nodes["all"].Order

	t.Logf("Execution order: deps=%d, build=%d, all=%d", depsOrder, buildOrder, allOrder)

	// Dependencies must run before dependents
	if depsOrder >= buildOrder {
		t.Errorf("deps (order %d) should run before build (order %d)", depsOrder, buildOrder)
	}
	if buildOrder >= allOrder {
		t.Errorf("build (order %d) should run before all (order %d)", buildOrder, allOrder)
	}

	// Specific check: deps should be 1, build should be 2, all should be 3
	if depsOrder != 1 {
		t.Errorf("deps should have order 1, got %d", depsOrder)
	}
	if buildOrder != 2 {
		t.Errorf("build should have order 2, got %d", buildOrder)
	}
	if allOrder != 3 {
		t.Errorf("all should have order 3, got %d", allOrder)
	}
}

// TestTopologicalSortDiamond tests parallel execution opportunities
// CONCEPT: In a diamond pattern, the middle nodes can run in parallel
//
//	  all
//	 /   \
//	A     B
//	 \   /
//	  deps
//
// Order should be: deps (1), then A and B (2), then all (3)
func TestTopologicalSortDiamond(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"deps"}},
		{Name: "B", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	if g.HasCycle {
		t.Fatal("Unexpected cycle in graph")
	}

	depsOrder := g.Nodes["deps"].Order
	aOrder := g.Nodes["A"].Order
	bOrder := g.Nodes["B"].Order
	allOrder := g.Nodes["all"].Order

	t.Logf("Execution order: deps=%d, A=%d, B=%d, all=%d", depsOrder, aOrder, bOrder, allOrder)

	// deps must run first
	if depsOrder != 1 {
		t.Errorf("deps should have order 1, got %d", depsOrder)
	}

	// A and B should run after deps but before all
	if aOrder <= depsOrder || aOrder >= allOrder {
		t.Errorf("A (order %d) should run after deps (%d) but before all (%d)",
			aOrder, depsOrder, allOrder)
	}
	if bOrder <= depsOrder || bOrder >= allOrder {
		t.Errorf("B (order %d) should run after deps (%d) but before all (%d)",
			bOrder, depsOrder, allOrder)
	}

	// A and B can run in parallel - they should have the same order
	// (or close order numbers, depending on how we count)
	t.Logf("A and B orders: %d and %d (can run in parallel)", aOrder, bOrder)
}

// TestTopologicalSortComplex tests a more complex dependency graph
// CONCEPT: Multiple paths, multiple levels
func TestTopologicalSortComplex(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build", "test"}},
		{Name: "build", Dependencies: []string{"compile"}},
		{Name: "test", Dependencies: []string{"compile"}},
		{Name: "compile", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	if g.HasCycle {
		t.Fatal("Unexpected cycle in graph")
	}

	// Print all orders for debugging
	t.Log("Execution orders:")
	for name, node := range g.Nodes {
		t.Logf("  %s: %d", name, node.Order)
	}

	// Verify dependency constraints
	deps := g.Nodes["deps"]
	compile := g.Nodes["compile"]
	build := g.Nodes["build"]
	test := g.Nodes["test"]
	all := g.Nodes["all"]

	// deps must run before compile
	if deps.Order >= compile.Order {
		t.Errorf("deps (%d) should run before compile (%d)", deps.Order, compile.Order)
	}

	// compile must run before build and test
	if compile.Order >= build.Order {
		t.Errorf("compile (%d) should run before build (%d)", compile.Order, build.Order)
	}
	if compile.Order >= test.Order {
		t.Errorf("compile (%d) should run before test (%d)", compile.Order, test.Order)
	}

	// build and test must run before all
	if build.Order >= all.Order {
		t.Errorf("build (%d) should run before all (%d)", build.Order, all.Order)
	}
	if test.Order >= all.Order {
		t.Errorf("test (%d) should run before all (%d)", test.Order, all.Order)
	}
}

// TestTopologicalSortIndependent tests independent targets
// CONCEPT: Targets with no dependencies or dependents can run in any order
func TestTopologicalSortIndependent(t *testing.T) {
	targets := []makefile.Target{
		{Name: "clean", Dependencies: nil},
		{Name: "lint", Dependencies: nil},
		{Name: "format", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// All should have order assigned
	for name, node := range g.Nodes {
		if node.Order == 0 {
			t.Errorf("Target %s should have an order assigned", name)
		}
		t.Logf("%s: order %d", name, node.Order)
	}

	// All are independent, so they should all have the same order (1)
	// since they can all run in parallel at the first level
	expectedOrder := 1
	for name, node := range g.Nodes {
		if node.Order != expectedOrder {
			t.Errorf("Independent target %s should have order %d, got %d",
				name, expectedOrder, node.Order)
		}
	}
}

// TestCriticalPathLinear tests critical path in a simple linear chain
// CONCEPT: In a linear chain, ALL nodes are critical (there's only one path)
func TestCriticalPathLinear(t *testing.T) {
	// Linear chain: all → build → compile → deps
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build"}},
		{Name: "build", Dependencies: []string{"compile"}},
		{Name: "compile", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// In a linear chain, ALL nodes should be critical
	for name, node := range g.Nodes {
		if !node.IsCritical {
			t.Errorf("Node %s should be critical in linear chain", name)
		}
	}

	t.Log("All nodes marked as critical (correct for linear chain)")
}

// TestCriticalPathDiamond tests critical path in a diamond pattern
// CONCEPT: The critical path goes through the longest branch
//
//	  all
//	 /   \
//	A     B (B has more dependencies, so it's critical)
//	|     |
//	|     C
//	 \   /
//	  deps
func TestCriticalPathDiamond(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"deps"}},
		{Name: "B", Dependencies: []string{"C"}},
		{Name: "C", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// The critical path is: all → B → C → deps (length 4)
	// A path is: all → A → deps (length 3) - NOT critical

	// Critical nodes should be: all, B, C, deps
	criticalNodes := []string{"all", "B", "C", "deps"}
	for _, name := range criticalNodes {
		if !g.Nodes[name].IsCritical {
			t.Errorf("Node %s should be critical", name)
		}
	}

	// A should NOT be critical (it's on the shorter path)
	if g.Nodes["A"].IsCritical {
		t.Error("Node A should NOT be critical (shorter path)")
	}

	t.Logf("Critical path correctly identified: all → B → C → deps")
}

// TestCriticalPathMultiplePaths tests when there are multiple critical paths
// CONCEPT: Two equally long paths = both are critical
//
//	    all
//	   /   \
//	  A     B
//	  |     |
//	  C     D
//	   \   /
//	    deps
func TestCriticalPathMultiplePaths(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"C"}},
		{Name: "B", Dependencies: []string{"D"}},
		{Name: "C", Dependencies: []string{"deps"}},
		{Name: "D", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// Both paths are equal length (4), so ALL nodes should be critical
	for name, node := range g.Nodes {
		if !node.IsCritical {
			t.Errorf("Node %s should be critical (both paths are equal)", name)
		}
	}

	t.Log("All nodes critical (both paths equal length)")
}

// TestParallelOpportunitiesDiamond tests parallel detection in diamond pattern
// CONCEPT: A and B can run in parallel (same level)
func TestParallelOpportunitiesDiamond(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"deps"}},
		{Name: "B", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// A and B should both be marked as parallelizable
	if !g.Nodes["A"].CanParallel {
		t.Error("A should be parallelizable")
	}
	if !g.Nodes["B"].CanParallel {
		t.Error("B should be parallelizable")
	}

	// deps and all are alone at their levels, so NOT parallel
	if g.Nodes["deps"].CanParallel {
		t.Error("deps should NOT be parallelizable (only one at its level)")
	}
	if g.Nodes["all"].CanParallel {
		t.Error("all should NOT be parallelizable (only one at its level)")
	}

	t.Log("Parallel opportunities correctly identified: A and B")
}

// TestParallelOpportunitiesComplex tests parallel detection in complex graph
func TestParallelOpportunitiesComplex(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build", "test", "lint"}},
		{Name: "build", Dependencies: []string{"compile"}},
		{Name: "test", Dependencies: []string{"compile"}},
		{Name: "lint", Dependencies: []string{"compile"}},
		{Name: "compile", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// build, test, and lint should all be parallelizable (same level)
	parallelTargets := []string{"build", "test", "lint"}
	for _, name := range parallelTargets {
		if !g.Nodes[name].CanParallel {
			t.Errorf("%s should be parallelizable", name)
		}
	}

	t.Log("Parallel opportunities: build, test, lint can run together")
}

// TestStandaloneTargetsNotMarked tests that standalone targets (no deps, no dependents)
// are NOT marked as critical or parallel since those concepts don't apply to them
// CONCEPT: Independent targets like "clean", "lint", "format" shouldn't get markers
func TestStandaloneTargetsNotMarked(t *testing.T) {
	targets := []makefile.Target{
		{Name: "clean", Dependencies: nil},
		{Name: "lint", Dependencies: nil},
		{Name: "format", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// None of these standalone targets should be marked as critical or parallel
	standaloneTargets := []string{"clean", "lint", "format"}
	for _, name := range standaloneTargets {
		node := g.Nodes[name]
		if node.IsCritical {
			t.Errorf("%s should NOT be marked as critical (it's just a standalone target)", name)
		}
		if node.CanParallel {
			t.Errorf("%s should NOT be marked as parallel (it's just a standalone target)", name)
		}
	}

	t.Log("Standalone targets correctly NOT marked as critical or parallel")
}

// TestNoParallelOpportunities tests when no parallelization is possible
// CONCEPT: Linear chain = no parallelization
func TestNoParallelOpportunities(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build"}},
		{Name: "build", Dependencies: []string{"compile"}},
		{Name: "compile", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	// In a linear chain, no node can run in parallel
	for name, node := range g.Nodes {
		if node.CanParallel {
			t.Errorf("Node %s should NOT be parallelizable (linear chain)", name)
		}
	}

	t.Log("No parallel opportunities in linear chain (correct)")
}
