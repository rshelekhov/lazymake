package graph

import (
	"strings"
	"testing"

	"github.com/rshelekhov/lazymake/internal/makefile"
)

// TestRenderTreeSimple tests basic tree rendering
func TestRenderTreeSimple(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build"}, Description: "Do everything"},
		{Name: "build", Dependencies: []string{"deps"}, Description: "Build the app"},
		{Name: "deps", Dependencies: nil, Description: "Install dependencies"},
	}

	g := BuildGraph(targets)

	renderer := TreeRenderer{
		ShowOrder:    true,
		ShowCritical: true,
		ShowParallel: true,
	}

	output := g.RenderTree(renderer)

	t.Log("Rendered tree:")
	t.Log(output)

	// Verify structure - should contain all target names
	requiredStrings := []string{"all", "build", "deps"}
	for _, req := range requiredStrings {
		if !strings.Contains(output, req) {
			t.Errorf("Output should contain %q", req)
		}
	}

	// Should contain tree characters
	if !strings.Contains(output, "└──") && !strings.Contains(output, "├──") {
		t.Error("Output should contain tree branch characters")
	}

	// Should contain descriptions
	if !strings.Contains(output, "Do everything") {
		t.Error("Output should contain target descriptions")
	}
}

// TestRenderTreeDiamond tests rendering with shared dependencies
// CONCEPT: deps appears twice but should only be fully rendered once
func TestRenderTreeDiamond(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"A", "B"}},
		{Name: "A", Dependencies: []string{"deps"}},
		{Name: "B", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	renderer := TreeRenderer{
		ShowOrder:    false,
		ShowCritical: false,
		ShowParallel: false,
	}

	output := g.RenderTree(renderer)

	t.Log("Rendered diamond tree:")
	t.Log(output)

	// Should contain "(see above)" for the second occurrence of deps
	if !strings.Contains(output, "(see above)") {
		t.Error("Should show '(see above)' for shared dependency")
	}

	// Should have proper tree structure
	lines := strings.Split(output, "\n")
	if len(lines) < 4 {
		t.Errorf("Expected at least 4 lines, got %d", len(lines))
	}
}

// TestRenderTreeWithAnnotations tests that all annotations appear correctly
func TestRenderTreeWithAnnotations(t *testing.T) {
	targets := []makefile.Target{
		{Name: "all", Dependencies: []string{"build", "test"}},
		{Name: "build", Dependencies: []string{"deps"}},
		{Name: "test", Dependencies: []string{"deps"}},
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	renderer := TreeRenderer{
		ShowOrder:    true,
		ShowCritical: true,
		ShowParallel: true,
	}

	output := g.RenderTree(renderer)

	t.Log("Rendered tree with annotations:")
	t.Log(output)

	// Should contain order numbers [1], [2], [3]
	if !strings.Contains(output, "[1]") {
		t.Error("Should contain execution order numbers")
	}

	// Should contain critical path marker ★
	if !strings.Contains(output, "★") {
		t.Error("Should contain critical path marker")
	}

	// Should contain parallel marker ∥ (build and test can run in parallel)
	if !strings.Contains(output, "∥") {
		t.Error("Should contain parallel marker")
	}
}

// TestRenderTreeCycle tests rendering when there's a circular dependency
func TestRenderTreeCycle(t *testing.T) {
	targets := []makefile.Target{
		{Name: "A", Dependencies: []string{"B"}},
		{Name: "B", Dependencies: []string{"C"}},
		{Name: "C", Dependencies: []string{"A"}}, // Cycle!
	}

	g := BuildGraph(targets)

	renderer := TreeRenderer{
		ShowOrder:    true,
		ShowCritical: true,
		ShowParallel: true,
	}

	output := g.RenderTree(renderer)

	t.Log("Rendered output for cycle:")
	t.Log(output)

	// Should show cycle warning
	if !strings.Contains(output, "Circular dependency") {
		t.Error("Should show circular dependency warning")
	}

	// Should show the cycle path
	if !strings.Contains(output, "→") {
		t.Error("Should show cycle path with arrows")
	}
}

// TestRenderTreeEmpty tests rendering an empty graph
func TestRenderTreeEmpty(t *testing.T) {
	var targets []makefile.Target

	g := BuildGraph(targets)

	renderer := TreeRenderer{}

	output := g.RenderTree(renderer)

	t.Log("Rendered empty graph:")
	t.Log(output)

	// Should have a message about no targets
	if !strings.Contains(output, "No targets") {
		t.Error("Should indicate no targets found")
	}
}

// TestRenderLegend tests legend generation
func TestRenderLegend(t *testing.T) {
	// Test with all options
	legend := RenderLegend(true, true, true)
	t.Logf("Full legend: %s", legend)

	if !strings.Contains(legend, "execution order") {
		t.Error("Legend should explain execution order")
	}
	if !strings.Contains(legend, "critical path") {
		t.Error("Legend should explain critical path")
	}
	if !strings.Contains(legend, "parallel") {
		t.Error("Legend should explain parallel")
	}

	// Test with no options
	emptyLegend := RenderLegend(false, false, false)
	if emptyLegend != "" {
		t.Errorf("Empty legend should be empty string, got %q", emptyLegend)
	}

	// Test with only one option
	orderOnly := RenderLegend(true, false, false)
	t.Logf("Order-only legend: %s", orderOnly)
	if !strings.Contains(orderOnly, "execution order") {
		t.Error("Should contain execution order explanation")
	}
	if strings.Contains(orderOnly, "critical") || strings.Contains(orderOnly, "parallel") {
		t.Error("Should NOT contain other explanations")
	}
}

// TestRenderTreeMultipleRoots tests rendering when there are multiple root nodes
// CONCEPT: Some Makefiles have multiple top-level targets
func TestRenderTreeMultipleRoots(t *testing.T) {
	targets := []makefile.Target{
		{Name: "build", Dependencies: []string{"deps"}},
		{Name: "test", Dependencies: []string{"deps"}},
		{Name: "clean", Dependencies: nil}, // Independent root
		{Name: "deps", Dependencies: nil},
	}

	g := BuildGraph(targets)

	renderer := TreeRenderer{
		ShowOrder:    true,
		ShowCritical: false,
		ShowParallel: false,
	}

	output := g.RenderTree(renderer)

	t.Log("Rendered tree with multiple roots:")
	t.Log(output)

	// Should contain all root targets
	if !strings.Contains(output, "build") {
		t.Error("Should contain 'build' root")
	}
	if !strings.Contains(output, "test") {
		t.Error("Should contain 'test' root")
	}
	if !strings.Contains(output, "clean") {
		t.Error("Should contain 'clean' root")
	}
}
