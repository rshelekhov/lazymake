package makefile

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

func ExpandPatternTargets(targets []Target, makefilePath string) ([]Target, error) {
	makefileDir := filepath.Dir(makefilePath)

	allTargets, err := getAllTargets(makefileDir)
	if err != nil {
		return targets, nil
	}

	var result []Target
	for _, target := range targets {
		if !target.IsPatternRule {
			result = append(result, target)
			continue
		}

		expanded := findMatchingTargets(allTargets, target.Name)
		if len(expanded) > 0 {
			for _, name := range expanded {
				result = append(result, Target{
					Name:          name,
					Description:   target.Description,
					CommentType:   target.CommentType,
					Dependencies:  target.Dependencies,
					Recipe:        target.Recipe,
					IsPatternRule: false,
				})
			}
		}
	}

	return result, nil
}

func findMatchingTargets(allTargets []string, patternTarget string) []string {
	if !strings.Contains(patternTarget, "%") {
		return nil
	}

	parts := strings.SplitN(patternTarget, "%", 2)
	prefix := parts[0]
	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
	}

	var result []string
	for _, t := range allTargets {
		if strings.Contains(t, "%") {
			continue
		}

		if !strings.HasPrefix(t, prefix) {
			continue
		}

		if suffix != "" && !strings.HasSuffix(t, suffix) {
			continue
		}

		result = append(result, t)
	}

	return result
}

func getAllTargets(makefileDir string) ([]string, error) {
	cmd := exec.Command("make", "-pn")
	cmd.Dir = makefileDir
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		if name == "" {
			continue
		}

		targets = append(targets, name)
	}

	return targets, nil
}
