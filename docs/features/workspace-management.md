# Workspace Management: Working with Multiple Projects

lazymake makes it easy to work with multiple projects and Makefiles. Press `w` to see recent workspaces and automatically discovered Makefiles in your project.

## Workspace Picker (Press `w`)

Access recent and discovered Makefiles with a single keypress:

```
┌─ Switch Workspace ────────────────────────────────────────┐
│                                                           │
│  FAVORITES                                                │
│  ⭐ Makefile                                              │
│     ./myapp • Last used: 2 minutes ago                    │
│                                                           │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    │
│                                                           │
│  ALL WORKSPACES                                           │
│  Makefile                                                 │
│  ../other-project • Last used: 1 hour ago                 │
│                                                           │
│  dangerous.mk                                             │
│  ./myapp/examples • Discovered                            │
│                                                           │
│  Makefile                                                 │
│  ./myapp/tools • Discovered                               │
│                                                           │
└───────────────────────────────────────────────────────────┘
  1 favorite • 3 workspaces
  enter: switch • f: favorite • esc/w: cancel
```

**Features:**
- **Automatic discovery**: Scans your project tree (up to 3 levels deep) to find all Makefiles
- **Recent workspaces**: Shows last 10 accessed Makefiles with directory path and last used time
- **Discovered workspaces**: Displays found Makefiles you haven't used yet
- **Favorites section**: Star frequently used projects with `f` - they appear in a dedicated section at the top
- **Visual separation**: Favorites and all workspaces are clearly separated with headers and dividers
- **Clear organization**: Filename in title, full relative path with root directory in description
- **Smart exclusions**: Skips `.git`, `node_modules`, `vendor`, build directories, and other common non-code paths
- **Fast scanning**: 5-second timeout ensures responsiveness even in large projects

## Status Bar Integration

The current workspace is always visible in the status bar:

```
└──────────────────────────────────────────────────────────┘
│ ./Makefile • 12 targets • 2 dangerous    enter: run • q  │
└──────────────────────────────────────────────────────────┘
```

The path is displayed relative to your current working directory:
- `./Makefile` - in current directory
- `../other/Makefile` - in sibling directory
- `~/projects/foo/Makefile` - absolute path with `~` expansion

## Per-Project History

Each workspace automatically maintains its own execution history. When you switch between projects, you'll see the recent targets for that specific Makefile:

```
# Working in project A
RECENT
⏱  build-api    Build the API server       3.2s
⏱  test-api     Run API tests              1.5s

# Switch to project B (press 'w')
RECENT
⏱  deploy-prod  Deploy to production       45.1s
⏱  build-web    Build web frontend         8.3s
```

This means:
- Each Makefile remembers its own frequently used targets
- No need to scroll through unrelated targets
- Faster context switching between projects

## Automatic Tracking

lazymake automatically tracks workspace usage:
- **On first use**: Creates workspace entry when you run a target
- **On subsequent uses**: Updates last accessed time
- **On cleanup**: Removes entries for deleted Makefiles automatically
- **Persistent**: Data survives across sessions in `~/.cache/lazymake/workspaces.json`

## How Discovery Works

When you press `w`, lazymake:
1. **Records current Makefile** - Ensures your current file appears in the list
2. **Scans project tree** - Searches up to 3 levels deep from current directory
3. **Finds all Makefiles** - Detects `Makefile`, `makefile`, `GNUmakefile`, `*.mk`, `*.mak`
4. **Applies exclusions** - Skips `.git`, `node_modules`, `vendor`, `build`, `dist`, `.cache`, etc.
5. **Combines results** - Shows recent workspaces first, then newly discovered ones
6. **Fast operation** - 5-second timeout prevents hanging on large projects

## Use Cases

### 1. Monorepo Development

Working with multiple Makefiles in a large repository:

```
my-monorepo/
├── Makefile              # Root Makefile
├── services/
│   ├── api/Makefile      # API service
│   ├── auth/Makefile     # Auth service
│   └── worker/Makefile   # Background worker
└── frontend/Makefile     # Frontend app
```

Press `w` to see all Makefiles automatically - no manual browsing needed!

### 2. Multi-Project Development

Switching between different projects:
- Press `w` to see recent projects and discovered Makefiles
- Star your most frequently used projects with `f`
- Favorites always appear at the top of the list

### 3. Onboarding

New to a project? Press `w`:
- Instantly see all available Makefiles
- Discovered Makefiles show "Discovered in project"
- Select one to start working - it gets added to your recent list

## For Makefiles Outside Discovery Range

If you need a Makefile that's:
- More than 3 levels deep
- In an excluded directory
- Outside your current project

Use the CLI flag:
```bash
lazymake -f path/to/Makefile
```

Once accessed, it appears in your recent workspaces list.

## Navigation

- **`w`**: Open workspace picker from list view
- **`↑/↓` or `j/k`**: Navigate workspaces
- **`f`**: Toggle favorite (star/unstar workspace)
- **`enter`**: Switch to selected workspace
- **`esc` or `w`**: Return to main list view

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
