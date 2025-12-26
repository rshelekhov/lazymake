# Keyboard Shortcuts Reference

Complete reference of all keyboard shortcuts in lazymake, organized by view.

## Main List View

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate up/down through targets |
| `j` / `k` | Vim-style navigation (up/down) |
| `Enter` | Execute the selected target |
| `g` | View dependency graph for selected target |
| `v` | Open variable inspector |
| `w` | Open workspace picker to switch Makefiles |
| `?` | Toggle help view (shows documented targets) |
| `/` | Enter search/filter mode |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Graph View

| Key | Action |
|-----|--------|
| `g` | Return to list view |
| `Esc` | Return to list view |
| `+` | Show more dependency levels (increase depth) |
| `=` | Show more dependency levels (alternate) |
| `-` | Show fewer dependency levels (decrease depth) |
| `_` | Show fewer dependency levels (alternate) |
| `o` | Toggle execution order numbers `[N]` |
| `c` | Toggle critical path markers `★` |
| `p` | Toggle parallel opportunity markers `||` |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Variable Inspector

| Key | Action |
|-----|--------|
| `v` | Return to list view |
| `Esc` | Return to list view |
| `↑` / `↓` | Navigate up/down through variables |
| `j` / `k` | Vim-style navigation (up/down) |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Output View

| Key | Action |
|-----|--------|
| `↑` / `↓` | Scroll through output |
| `j` / `k` | Vim-style scrolling (up/down) |
| `Esc` | Return to list view |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Workspace Picker

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate up/down through workspaces |
| `j` / `k` | Vim-style navigation (up/down) |
| `Enter` | Switch to selected workspace |
| `f` | Toggle favorite (star/unstar) for selected workspace |
| `w` | Return to list view (cancel) |
| `Esc` | Return to list view (cancel) |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Help View

| Key | Action |
|-----|--------|
| `?` | Return to list view |
| `Esc` | Return to list view |
| `q` | Quit lazymake |
| `Ctrl+C` | Quit lazymake |

## Search/Filter Mode

| Key | Action |
|-----|--------|
| Type characters | Filter targets by name or description |
| `Backspace` | Delete last character from search query |
| `Esc` | Clear search and return to full list |
| `Enter` | Execute selected filtered target |
| `↑` / `↓` | Navigate filtered results |
| `j` / `k` | Vim-style navigation through filtered results |

## Tips

- **Vim-style navigation**: All views support both arrow keys and `j/k` for navigation
- **Universal quit**: `q` or `Ctrl+C` works in any view to exit lazymake
- **Context-sensitive help**: The status bar always shows relevant shortcuts for the current view
- **ESC key**: Generally returns to the previous view or cancels the current action

---

[← Back to Documentation](../README.md) | [← Back to Main README](../../README.md)
