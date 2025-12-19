# Syntax Highlighting for Multi-Language Recipes

lazymake automatically detects and syntax-highlights recipes written in different programming languages, making it easier to read and understand complex build commands.

## Visual Preview

When viewing a target, recipes are automatically highlighted based on their detected language:

```
┌─────────────────────┬────────────────────────────────────────┐
│ ALL TARGETS         │ build:                                 │
│ > docker-build      │                                        │
│   python-test       │   docker build -t myapp:latest .      │
│   npm-install       │   docker push myapp:latest            │
│   go-build          │   docker images | grep myapp          │
│                     │                                        │
│                     │   [dockerfile]                         │
│                     │                                        │
│                     │   💡 Press 'g' to view full graph      │
└─────────────────────┴────────────────────────────────────────┘
```

Keywords, strings, comments, operators, and other syntax elements are colored according to their type, using a monokai-inspired color scheme optimized for terminal readability.

## How It Works

### Automatic Language Detection

lazymake uses multiple heuristics to detect the language of your recipes:

**1. Shebang Detection (Highest Priority)**

```makefile
python-script:
	#!/usr/bin/env python3
	import sys
	print(f"Python version: {sys.version}")
```

The shebang line `#!/usr/bin/env python3` tells lazymake to highlight this as Python.

**2. Command Pattern Matching**

```makefile
docker-build:
	docker build -t myapp .
	docker run myapp
```

Commands like `docker`, `npm`, `go`, `cargo`, etc. are recognized and the appropriate lexer is applied.

**3. Fallback to Bash**

If no specific language is detected, recipes default to bash/shell highlighting:

```makefile
generic-task:
	echo "Hello, World!"
	ls -la
```

### Manual Language Override

You can explicitly specify a language using a comment above the target:

```makefile
# language: python
generate-config:
	config = {
		"version": "1.0.0",
		"debug": True
	}
	print(json.dumps(config))
```

Supported comment formats:
- `# language: <lang>`
- `# lang: <lang>`
- `# syntax: <lang>`

## Supported Languages

lazymake supports **100+ programming languages** via the Chroma syntax highlighting library. Common languages include:

### Development Languages
- **Go**: `go build`, `go test`, `go run`
- **Python**: `python`, `python3`, `pip`, `poetry`
- **JavaScript/TypeScript**: `npm`, `yarn`, `node`, `npx`
- **Rust**: `cargo build`, `cargo test`, `cargo run`
- **C/C++**: `gcc`, `g++`, `clang`, `make`, `cmake`
- **Java**: `javac`, `java`, `mvn`, `gradle`
- **Ruby**: `ruby`, `bundle`, `gem`
- **PHP**: `php`, `composer`

### DevOps & Infrastructure
- **Dockerfile**: Dockerfile syntax (`FROM`, `RUN`, `COPY`) - use `# language: docker` for embedded Dockerfiles
- **Kubernetes**: `kubectl`, `helm` (use `# language: yaml` for manifests)
- **Shell**: `bash`, `sh`, and shell commands (default for most Makefile recipes)

### Language Aliases

For convenience, lazymake recognizes common aliases:
- `py` → `python`
- `js` → `javascript`
- `golang` → `go`
- `sh`/`shell` → `bash`
- `yml` → `yaml`
- `dockerfile` → `docker`

## Color Scheme

lazymake uses a monokai-inspired color palette designed for readability in terminal environments:

| Element | Color | Hex | Example |
|---------|-------|-----|---------|
| Keywords | Pink | `#FF79C6` | `def`, `if`, `for`, `func` |
| Strings | Yellow | `#E6DB74` | `"hello"`, `'world'` |
| Comments | Gray | `#75715E` | `# comment`, `// comment` |
| Functions | Green | `#A6E22E` | `print()`, `build()` |
| Numbers | Purple | `#AE81FF` | `42`, `3.14` |
| Operators | Red | `#F92672` | `+`, `-`, `=`, `\|` |
| Variables | Orange | `#FD971F` | `$VAR`, `${VAR}` |
| Types | Cyan | `#66D9EF` | `int`, `string` |
| Default | White | `#F8F8F2` | Everything else |

These colors:
- Work well on both dark and light terminal backgrounds
- Provide sufficient contrast for readability
- Match common editor themes (VS Code, Sublime Text)
- Degrade gracefully on limited-color terminals

## Examples

### Python with Shebang

```makefile
## Run Python data processing script
process-data:
	#!/usr/bin/env python3
	import pandas as pd

	def process():
		df = pd.read_csv("data.csv")
		result = df.groupby("category").sum()
		print(result)

	process()
```

**Detected as:** Python
**Highlights:** `import`, `def`, `return` (keywords), `"data.csv"` (string), `# comments`

### Docker CLI Commands

```makefile
## Build and push Docker image
docker-deploy:
	docker build -t myapp:latest .
	docker tag myapp:latest myapp:$(VERSION)
	docker push myapp:latest
	docker push myapp:$(VERSION)
```

**Detected as:** Bash (default)
**Highlights:** Shell syntax, commands, strings, variables

### Embedded Dockerfile

```makefile
## Generate Dockerfile
# language: docker
create-dockerfile:
	cat > Dockerfile << 'EOF'
	FROM golang:1.21-alpine
	WORKDIR /app
	COPY . .
	RUN go build -o /app/myapp
	CMD ["/app/myapp"]
	EOF
```

**Detected as:** Docker (manual override)
**Highlights:** Dockerfile keywords (`FROM`, `RUN`, `COPY`), image names, paths

### Go Build with Flags

```makefile
## Build Go binary with version info
build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o bin/app \
		./cmd/app
	go test -v -race ./...
```

**Detected as:** Go
**Highlights:** `go` commands, flags, paths

### Complex Shell Script

```makefile
## Deploy application with health checks
deploy:
	#!/bin/bash
	set -euo pipefail

	echo "Deploying version $(VERSION)..."

	for i in {1..3}; do
		if curl -f http://localhost:8080/health; then
			echo "Deployment successful!"
			exit 0
		fi
		echo "Retry $i/3..."
		sleep 5
	done

	echo "Deployment failed!"
	exit 1
```

**Detected as:** Bash (from shebang)
**Highlights:** `set`, `echo`, `for`, `if` (keywords), variables, strings

### Manual Override for Embedded Languages

```makefile
## Generate Kubernetes manifests
# language: yaml
k8s-manifests:
	cat > deployment.yaml << 'EOF'
	apiVersion: apps/v1
	kind: Deployment
	metadata:
	  name: myapp
	spec:
	  replicas: 3
	  template:
	    spec:
	      containers:
	      - name: app
	        image: myapp:latest
	EOF
	kubectl apply -f deployment.yaml
```

**Detected as:** YAML (manual override)
**Highlights:** YAML structure, keys, values

## Performance

Syntax highlighting is optimized for performance:

- **Caching**: Highlighted code is cached with an LRU cache (1000 entries)
- **First render**: < 5ms per recipe
- **Cached render**: < 0.5ms per recipe
- **Memory usage**: ~2-5 MB for cache (negligible)

The cache ensures that navigating between targets feels instant, even with complex recipes.

## Detection Priority

When multiple detection methods are available, lazymake uses this priority order:

1. **Manual override** (`# language: <lang>`) - Always wins
2. **Shebang detection** (`#!/usr/bin/env <interpreter>`) - Very reliable
3. **Command pattern matching** - Good heuristic with weighted voting
4. **Fallback to bash** - Safe default for Makefile recipes

### Example: Priority in Action

```makefile
# language: python
mixed-commands:
	#!/bin/bash
	docker build -t test .
	echo "Building..."
```

**Result:** Highlighted as **Python** (manual override wins over shebang and docker commands)

## Language Badge

When a non-bash language is detected, lazymake displays a language badge below the recipe:

```
build:
  go build -o app ./cmd/app
  go test ./...

  [go]

  💡 Press 'g' to view full dependency graph
```

This helps you quickly identify what language the recipe is written in.

## Tips & Best Practices

### 1. Use Shebangs for Scripts

For multi-line scripts, always include a shebang:

```makefile
install-deps:
	#!/usr/bin/env python3
	import subprocess
	subprocess.run(["pip", "install", "-r", "requirements.txt"])
```

### 2. Use Manual Overrides for Embedded Code

When generating code or using heredocs, specify the language:

```makefile
# lang: javascript
generate-package:
	cat > package.json << 'EOF'
	{
	  "name": "myapp",
	  "version": "1.0.0",
	  "scripts": {
	    "build": "webpack"
	  }
	}
	EOF
```

### 3. Document Complex Recipes

Use `##` comments to explain what multi-language recipes do:

```makefile
## Build Python wheel and upload to PyPI
# language: python
publish:
	#!/usr/bin/env python3
	import os
	import subprocess

	subprocess.run(["python", "setup.py", "bdist_wheel"])
	subprocess.run(["twine", "upload", "dist/*"])
```

### 4. Test Different Terminals

Colors may appear different across terminal emulators. The color scheme is designed to work well with:
- iTerm2 (macOS)
- Terminal.app (macOS)
- Alacritty
- GNOME Terminal
- Windows Terminal
- VS Code integrated terminal

## Troubleshooting

### Language Not Detected Correctly

**Problem:** lazymake highlights your recipe as bash when it's actually Python.

**Solution:** Add a manual language override:

```makefile
# language: python
my-target:
	print("Hello, World!")
```

### Colors Look Wrong

**Problem:** Colors don't match the documentation or look incorrect.

**Possible causes:**
1. Terminal doesn't support true color (24-bit color)
2. Terminal color scheme conflicts with highlighting

**Solution:** Use a terminal emulator with true color support (most modern terminals). Check with:

```bash
echo $COLORTERM  # Should output "truecolor" or "24bit"
```

### Highlighting Too Slow

**Problem:** Noticeable lag when switching between targets.

**This is very rare**, but if it happens:
- You may have extremely long recipes (>1000 lines)
- Cache may be disabled or not working

lazymake automatically skips highlighting for recipes over 100 lines and shows a warning.

## Examples in Action

Try the included example Makefile to see all supported languages:

```bash
lazymake -f examples/highlighting.mk
```

This demonstrates:
- Automatic detection for 10+ languages
- Manual overrides
- Shebang handling
- Complex multi-line recipes
- Mixed-language targets

## Technical Details

### Implementation

- **Syntax highlighter**: [Chroma v2](https://github.com/alecthomas/chroma) (supports 200+ languages)
- **Terminal styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) (Charm Bracelet)
- **Cache**: Custom LRU cache with SHA-256 key generation
- **Detection**: Regex-based pattern matching with weighted voting

### Language Lexers

Chroma provides lexers for all major programming languages. When you specify a language (manually or via detection), lazymake:

1. Loads the appropriate Chroma lexer
2. Tokenizes the code
3. Maps token types to Lipgloss colors
4. Renders the styled output
5. Caches the result

For unknown languages, lazymake falls back to the bash lexer, which handles most shell-like syntax reasonably well.

## Related Features

- [Variable Inspector](variable-inspector.md) - See what variables your recipes use
- [Dependency Graph](dependency-graphs.md) - Visualize build dependencies
- [Safety Features](safety-features.md) - Catch dangerous commands before execution

---

**See it in action:** Run `lazymake -f examples/highlighting.mk` to explore syntax highlighting with real examples!
