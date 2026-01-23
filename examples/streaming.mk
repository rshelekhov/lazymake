# Makefile for testing streaming output feature
# Usage: lazymake -f examples/streaming.mk

## Slow target with incremental output
slow:
	@echo "Starting slow task..."
	@sleep 1
	@echo "Step 1 complete"
	@sleep 1
	@echo "Step 2 complete"
	@sleep 1
	@echo "Step 3 complete"
	@sleep 1
	@echo "All done!"

## Rapid output (many lines quickly)
rapid:
	@for i in $$(seq 1 50); do echo "Line $$i"; done

## Mixed timing output
mixed:
	@echo "Fast line 1"
	@echo "Fast line 2"
	@sleep 2
	@echo "After 2 second pause"
	@echo "Fast line 3"
	@sleep 1
	@echo "After 1 second pause"
	@echo "Done"

## Long running with progress simulation
progress:
	@echo "Processing..."
	@for i in $$(seq 1 10); do \
		echo "Progress: $$i/10"; \
		sleep 0.5; \
	done
	@echo "Complete!"

## Target with no output
silent:
	@sleep 3

## Target that fails mid-execution
failing:
	@echo "Starting..."
	@sleep 1
	@echo "About to fail..."
	@sleep 1
	@exit 1

## Verbose output with timestamps
verbose:
	@echo "[`date +%H:%M:%S`] Starting verbose task"
	@sleep 1
	@echo "[`date +%H:%M:%S`] Connecting to service..."
	@sleep 1
	@echo "[`date +%H:%M:%S`] Fetching data..."
	@sleep 1
	@echo "[`date +%H:%M:%S`] Processing results..."
	@sleep 1
	@echo "[`date +%H:%M:%S`] Task completed"

## Scrollable output (more lines than viewport)
scrollable:
	@for i in $$(seq 1 100); do \
		echo "Log entry $$i: This is a sample log message for testing viewport scrolling"; \
		sleep 0.1; \
	done
