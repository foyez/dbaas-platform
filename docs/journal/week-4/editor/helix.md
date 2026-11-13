# Helix Editor — Cheat Sheet & Setup Guide

> For Go, TypeScript, React/TSX, and C++ — coming from VS Code / NvChad / Neovim.
> Helix ships with file navigation, fuzzy pickers, LSP, Tree-sitter, diagnostics, and multi-selection built in. You generally don't need plugins to recreate your old setup.

**Core mental model:** Helix is *selection-oriented*. Instead of `move → command`, think `select → command`.

---

## 1. Install & Verify

```bash
brew install helix
hx --version
hx .                    # open current dir as project

hx --health              # check LSP health
hx --health go
hx --health typescript
hx --health tsx
hx --health cpp
```

---

## 2. Keybinding Map: Helix ↔ NvChad/Neovim ↔ VS Code

| Action | Helix | NvChad / Neovim | VS Code |
|---|---|---|---|
| Open project | `hx .` | `nvim .` | `code .` |
| File explorer | `Space e` | NvimTree | `Cmd/Ctrl+Shift+E` |
| Find file | `Space f` | Telescope find_files | `Cmd/Ctrl+P` |
| Open buffers/tabs | `Space b` | Telescope buffers | `Cmd/Ctrl+P` (tabs) |
| File symbols | `Space s` | Telescope symbols | `Cmd/Ctrl+Shift+O` |
| Workspace symbols | `Space S` | Telescope ws symbols | `Cmd/Ctrl+T` |
| File diagnostics | `Space d` | Trouble/LSP diag | `Cmd/Ctrl+Shift+M` |
| Workspace diagnostics | `Space D` | Trouble | Problems panel |
| Go to definition | `gd` | `gd` | `F12` |
| Go to references | `gr` | `gr` | `Shift+F12` |
| Go to implementation | `gi` | `gi` | `Ctrl+F12` |
| Go to type definition | `gy` | — | `Ctrl+Shift+F12` |
| Hover / docs | `Space k` | `K` | Hover on symbol |
| Rename symbol | `Space r`* | `Space rn` | `F2` |
| Insert mode | `i` | `i` | (always insert) |
| Normal mode | `Esc` | `Esc` | — |
| Select/Visual mode | `v` | `v` (Visual) | Mouse/Shift+arrows |
| Undo | `u` | `u` | `Cmd/Ctrl+Z` |
| Redo | `U` | `Ctrl-r` | `Cmd/Ctrl+Shift+Z` |
| Delete line (select+op) | `x` then `d` | `dd` | `Cmd/Ctrl+Shift+K` |
| Copy line (select+op) | `x` then `y` | `yy` | `Cmd/Ctrl+C` |
| Paste | `p` | `p` | `Cmd/Ctrl+V` |
| Search | `/` | `/` | `Cmd/Ctrl+F` |
| Search & replace | `:%s/foo/bar/g` | `:%s/foo/bar/g` | `Cmd/Ctrl+H` |
| Save | `:w` | `:w` | `Cmd/Ctrl+S` |
| Quit | `:q` | `:q` | `Cmd/Ctrl+W` |
| Save & quit | `:wq` | `:wq` | — |
| Format file | `:format` | `:Format`/LSP | `Shift+Alt+F` |
| Vertical split | `Ctrl-w v` | `Ctrl-w v` | `Cmd/Ctrl+\` |
| Horizontal split | `Ctrl-w s` | `Ctrl-w s` | `Cmd/Ctrl+Shift+\` |
| Move between splits | `Ctrl-w h/j/k/l` | `Ctrl-w h/j/k/l` | `Cmd/Ctrl+1/2/3` |
| Jump back / forward | `Ctrl-o` / `Ctrl-i` | `Ctrl-o` / `Ctrl-i` | `Alt+←` / `Alt+→` |

\* Confirm your `Space r` rename binding with `hx --health` or the `Space` menu — mappings can vary by version.

> ⚠️ **Prefer LSP rename over search/replace** for renaming variables, functions, or types — the language server understands the symbol scope; text substitution doesn't.

---

## 3. Navigation Basics

| Keys | Action |
|---|---|
| `h j k l` | left / down / up / right |
| `w` / `b` / `e` | next/prev/end of word |
| `W` / `B` / `E` |next/previous/end of WORD |
| `42gg` / `42G` | go to line 42 |
| `Ctrl-u` / `Ctrl-d` | half-page up / down |
| `Ctrl-b` / `Ctrl-f` | full page up / down |
| `Ctrl-o` / `Ctrl-i` | jump back / forward |

### `g` Navigation

| Keys | Action |
|---|---|
| `gg` / `ge` | start/end of file |
| `gh` / `gl` | start/end of line |
| `gt` / `gc` / `gb` |top/center/bottom of screen|
| `gd` / `gD` / `gr` / `gi` / `gy` | go to definition/declaration/references/implementation/type definition |
|`ga`| last accessed/alternate file |
|`gm`| last modified/alternate file |
|`gn` / `gp`| next / previous buffer |

## 4. Editing Basics

| Keys | Action |
|---|---|
| `x` | select current line |
| `d` | delete selection |
| `c` | change selection |
| `y` | yank (copy) selection |
| `p` / `P` | paste after / before |
| `J` | join lines |
| `u` / `U` | undo / redo |

**Common pattern — duplicate line below:**
`x` (select line) → `y` (copy) → `j` (move down) → `p` (paste)

## 5. Modes

| Key | Enters |
|---|---|
| `i` / `a` | insert before / after cursor |
| `I` / `A` | insert at line start / end |
| `o` / `O` | new line below / above |
| `v` | select mode |
| `Esc` | back to normal mode |

## 6. Search & Replace

| Keys | Action |
|---|---|
| `/` / `?` | search forward/backward |
| `n` / `N` | next / previous search match |
| `*` | search word under cursor |
| `:%s/foo/bar/g` | replace all |
| `:%s/foo/bar/gc` | replace with confirmation |

## 7. File Commands

| Command | Action |
|---|---|
| `:w` | save |
| `:q` | quit |
| `:wq` | save + quit |
| `:q!` | quit without saving |
| `:wa` / `:qa` / `:wqa` | save/quit/save+quit all |

---

## 8. Language Servers to Install

| Language | Server | Install |
|---|---|---|
| Go | `gopls` | `brew install go && go install golang.org/x/tools/gopls@latest` |
| TS / JS / React | `typescript-language-server` | `brew install node && npm install -g typescript typescript-language-server` |
| C / C++ | `clangd` | `brew install llvm` |

**Formatters:**
- Go → `gofmt` (built-in, run via `:format`)
- TS/React → `prettier` (`npm install --save-dev prettier`)
- C++ → `clang-format` (comes with LLVM)

---

## 9. Minimal Starter Config

`~/.config/helix/config.toml`:

```toml
theme = "catppuccin_mocha"

[editor]
line-number = "relative"
cursorline = true
true-color = true
color-modes = true
bufferline = "multiple"
auto-pairs = true

[editor.cursor-shape]
insert = "bar"
normal = "block"
select = "underline"

[editor.indent-guides]
render = true
character = "│"

[editor.soft-wrap]
enable = false
```

Test a theme without saving: `:theme catppuccin_mocha`

**Config structure:**
```
~/.config/helix/
├── config.toml       # editor settings, theme, keys
└── languages.toml    # LSPs, formatters, per-language config
```
Project-local overrides go in `.helix/config.toml` / `.helix/languages.toml`.

---

## 10. Daily Workflow Cheats by Language

**Go:** `Space f` find → `gd` def → `gr` refs → `Space d` diagnostics → `:format`  
**React/TS:** `Space f` find component → `Space s` symbols → `gd`/`gr` → `Space k` docs  
**C++:** `gd` def → `gi` implementation → `gy` type def → `Space k` docs  

---

## 11. 25 Commands to Memorize First

```
Modes:     i  Esc  v
Movement:  h j k l   w b e   gg  ge
Editing:   x  d  c  y  p  u  U
Search:    /  n  N  *
Files:     Space e   Space f   Space b   Space s
LSP:       gd  gr  gi  gy   Space k   Space d   Space D
Commands:  :w  :q  :wq  :q!
```

## 12. Buffer Commands

```
:bc — close current buffer
:bco — close all other buffers
:bn — next buffer
:bp — previous buffer
:buffer-close — same as :bc
```

---

## Final Notes

- Don't recreate your Neovim config in Helix — use the built-ins first (picker, explorer, LSP, Tree-sitter, diagnostics, multi-select).
- Keep `config.toml` minimal; only add settings once something actually annoys you.
- The one concept that unlocks Helix: **SELECT → OPERATE**, not MOVE → OPERATE.