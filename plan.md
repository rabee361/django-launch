# django-launch: current state and what is left

## where things stand

The suite is green. `go test ./...` passes in all three internal packages, `gofmt -l .` is empty, `go vet ./...` is clean, and `go build ./...` succeeds. Coverage is 100.0% in `internal/config`, 88.5% in `internal/generator`, and 37.5% in `internal/tui`, across 32 test functions. The end-to-end test still needs Python or uv plus network, everything else runs offline.

Nothing is committed. `git status` shows 17 modified files, 5 untracked ones (`README.md`, `LICENSE`, `.gitignore`, `.github/`, `plan.md`), and two staged deletions (`ascii_art.txt`, `django-launch.exe`).

## what is done

### Wizard behavior

- Space toggles a dependency row. `update.go:129` matches `case "space", "x"`, which is what `Key.String()` actually returns for that key (ultraviolet `key.go:444-464`).
- Paste works. `update.go:36-41` routes `tea.PasteMsg` to `m.textInput.Update` on step 1 and drops it on the other steps, so Ctrl+V inserts text up to the 50 character limit. `TestUpdate_Paste` covers both paths.
- The cursor work landed. Step 1 uses the real terminal cursor via `SetVirtualCursor(false)` at `model.go:70`, with `Y` corrected from the row count in `views.go:33-70`. Step 3 shows a blinking bar at `CursorBar`, and every other step hides the cursor.
- The placeholder renders in full. `ti.SetWidth(50)` at `model.go:73` fixes the one-letter display, because `placeholderView` sizes its buffer from `Width()`.
- The banner is a string literal at `views.go:15-29` and `ascii_art.txt` is deleted, so `View()` no longer reads the disk and the binary no longer depends on the directory it was launched from.

### Help text and README

- The hints at `views.go:88`, `views.go:122`, `views.go:158`, `views.go:201`, `views.go:251`, and `views.go:265` now say what the code does, and the README keys table says the same thing. Both cover `y` and `n` advancing on step 2, the `x` and `backspace` aliases, `b`, and `esc` on the done and error screens. The confirm prompt line matches too.
- The README gained a `generating` row, a paste note on step 1, and `left`/`right` on step 2. The stale sentence about the hint not mentioning `y` and `n` is gone.
- The two are still separate literals in `views.go` and `README.md`, so they can drift again. The structural fix sits in item 9.

### Repo files

- `README.md` has the project layout, requirements, build steps, the four steps, a keys table written from the code, and the optional package table.
- `LICENSE` is MIT under `Copyright (c) 2026 rabee361`. `.gitignore` covers binaries, profiles, `__pycache__`, editors, and OS files. `.github/workflows/ci.yml` runs gofmt, vet, `go mod tidy -diff`, build, govulncheck, and a three-OS test matrix, with the automatic triggers commented out so nothing runs until you uncomment them.
- The 5 MB `django-launch.exe` is staged for removal and matches `*.exe` in `.gitignore`.

### Fixes and structure

- Both previously failing tests pass. `model_test.go` expects 9 options with the 9 IDs in `model.go` order, and the end-to-end fixture passes `djangorestframework`, which its `settings.py` assertion was checking for.
- The module path is `github.com/rabee361/django-launch`. `go.mod:1` and all 11 import sites changed, so the module can be imported and `go mod tidy` stays happy.
- `go mod tidy` has been run, so the three direct dependencies are no longer marked `// indirect`.

## what is left

Ordered by what a user notices first.

### User-visible

1. **No resize handling.** `WindowSizeMsg` is never read, so nothing adapts when the terminal changes size. It also matters for the cursor math, because rows are counted as newlines in the frame string while the terminal wraps long lines. On a narrow terminal the bar lands one row off for every line that wrapped. Store `m.width`, cap the description column with it, and the two problems resolve together. *Half a day.*

2. **Generation cannot be cancelled.** `StepExecuting` has no key case at all, so a slow `uv pip install` has no escape short of killing the terminal. Ctrl+C quits the program and leaves the generator goroutine parked forever on the unbuffered sends at `model.go:188` and `model.go:194`. Nothing breaks today because the process exits anyway, and that is the problem. The first time you add cleanup on exit or a confirmation before quitting, it leaks. A `context.Context` threaded through `GenerateProject` fixes both the cancel key and the leak. *Half a day.*

3. **Inline output has not been decided.** `main.go:13` calls `tea.NewProgram(tui.NewModel()).Run()` with no options, so the wizard renders inline and scrolls with the terminal's history. Alt-screen is the usual choice for a multi-step TUI because it restores the scrollback on exit. Pick one and say which in the README. *10 minutes.*

### Under the hood

4. **Errors are unwrappable in name only.** Three places wrap with `%v` instead of `%w`, `venv.go:64`, `venv.go:106`, and `django.go:21`. `errors.Is` on the underlying `exec.ExitError` never matches, which is what a "uv is not installed, here is how to get it" hint would need. The wraps in `generator.go` already use `%w`. *10 minutes.*

5. **The spinner never stops ticking.** `update.go:43-46` re-arms `spinner.Tick` on every step, including Done and Error, until the process quits. Guard it on `m.step == StepExecuting`. *10 minutes.*

6. **Error state is left dirty.** `update.go:55-59` sets `StepError` without clearing `progressItems` or `currentTask`. `viewError` does not render them today, so this is a trap waiting for whoever next adds a "what went wrong" list. *10 minutes.*

7. **`progressMsg.step` and `.total` are written and never read.** They are set at `model.go:188-190` and ignored by `update.go:48-53`, which only consumes `text`. Either show `3/6` in the executing view, which the data already supports, or delete the fields. Showing it is the better of the two, since the generator already reports them. *10 minutes.*

### Code quality

8. **Dead field.** `Model.pythonCmd` (`model.go:64`, written at `model.go:90`) is never read. `cfg.PythonCmd` is set alongside it at `model.go:87` and already holds the same value. *10 minutes.*

9. **View tests, then real key bindings.** `internal/tui` sits at 37.5% and `views.go` contributes none of it. Golden tests that render each step and compare against an expected string would protect the cursor row math, the help text, and every layout change. They are also what makes the next step safe. Move the keys into `bubbles/help` with `key.Binding` and render each hint from the bindings, so the hints, the confirm prompt, and the README table cannot drift apart again. The space bug was exactly what happens without that. *Half a day.*

### Release

10. **Enable CI.** The workflow is in place and the suite is green, so this is a decision rather than a task. Uncomment the `push` and `pull_request` triggers in `.github/workflows/ci.yml` when you are ready for every push to be gated. *10 minutes.*

11. **Version and releases.** Add `--version` backed by `-ldflags`, then GoReleaser for Windows, Linux, and macOS archives. A bug report that names a version is actionable. *Half a day.*

12. **CHANGELOG and issues.** Keep a Changelog format for users, and move `TODO.md`'s three items (database selection, caching plus django-redis, email settings) into issues so they have states. *10 minutes.*

## layout: lipgloss versus writing by hand

Kept from the previous plan. The line references have moved, the advice has not.

### The renderer owns the terminal

bubbletea keeps a cell buffer of the frame it last painted. On each pass it compares the new frame against that buffer, queues only the cells that changed, hides the cursor, flushes the batch as one write, and then moves the cursor to the X and Y carried by `tea.View.Cursor`. Every byte that reaches the screen is supposed to come from that one writer.

A `fmt.Print` or an `os.Stdout.Write` inside `Update` or `View` bypasses it. Three things go wrong.

- Your bytes appear immediately while the renderer's buffer still holds the old frame.
- The next diff compares the new frame against that stale buffer. Cells where the two agree are skipped, so they keep your printed text. Cells where they disagree get repainted from the frame. The screen ends up stitched together from two writers, and the tear stays up until something forces a full repaint, such as a resize.
- `Update` runs on the event loop goroutine while the renderer flushes from its own. The two writes can interleave inside an escape sequence, leaving the terminal in a color state no frame ever set.

The program runs inline, because `main.go:13` passes no options. A print that ends in a newline scrolls the terminal, so every row moves up while the renderer still believes row zero sits where it did. Every later diff paints off by that scroll.

The cursor math breaks the same way. `View` counts rows inside its own string.

```go
base := strings.Count(b.String(), "\n")
...
s, row := m.viewProjectName()
nameRow = base + row
```

`nameRow` counts from the top of the frame that string describes. A print adds rows the string never contained, so `tea.Cursor` with `Y = nameRow` points at a different screen row than the text input. One stray line and the blinking bar sits on the help text instead of in the input.

If a message has to leave the program, use the renderer's own channel. `tea.Println` and `tea.Printf` return a `Cmd`, print above the frame, and the renderer shifts its bookkeeping for them. Everything else belongs in the string `View` returns.

### Measuring width against styled text

The current style builds one string with `strings.Builder` and one `fmt.Sprintf` per line. The dependency row is two writes, at `views.go:153-154`.

```go
b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(opt.Name)))
b.WriteString(fmt.Sprintf("      %s\n", subtitleStyle.Render(opt.Description)))
```

The six spaces are correct today because `cursor` is two cells, `checkbox` is three, and one space separates them. Nothing in this file calls `len()` on a styled string, which is the only reason the columns hold.

The first width calculation you add will be wrong. `style.Render` and `uncheckedStyle.Render("[ ]")` return text carrying SGR sequences. `len()` counts those bytes. `%*s` pads to that count. A nineteen character package name with bold green applied measures in the low forties instead of nineteen, so a padded name column drifts right by the escape bytes and the rows in the list stagger. `lipgloss.Width` reads the same string with an ANSI aware measure and returns nineteen.

The confirm card at `views.go:189-196` has the same trap, with the labels plain and the values styled.

```go
cardContent := fmt.Sprintf(
	"Project Name : %s\nTarget Path  : ./%s\nDocker       : %s\nDependencies : Django, %s\nEnvironment  : %s",
	selectedItemStyle.Render(m.cfg.ProjectName),
	m.cfg.ProjectName,
	dockerText,
	depsText,
	secondaryColorStyle(toolText),
)
```

The hand counted spaces after each label work because labels carry no escapes. Any attempt to measure the card, to pad the value column, or to check the card fits the terminal runs into `selectedItemStyle.Render` and `secondaryColorStyle(toolText)`.

`JoinHorizontal` fixes the row by measuring for you. It splits each block into lines, takes each block's widest line with an ANSI aware measure, and pads the shorter block with spaces.

Before:

```go
marker := cursor + checkbox + " "
b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, style.Render(opt.Name)))
b.WriteString(fmt.Sprintf("      %s\n", subtitleStyle.Render(opt.Description)))
```

After:

```go
marker := cursor + checkbox + " "
gutter := lipgloss.JoinVertical(lipgloss.Left,
	marker,
	strings.Repeat(" ", lipgloss.Width(marker)))
body := lipgloss.JoinVertical(lipgloss.Left,
	style.Render(opt.Name),
	subtitleStyle.Render(opt.Description))
b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, gutter, body))
b.WriteString("\n")
```

Two lines go in and two lines come out, so `strings.Count` still sees two newlines per row. The indent is now derived from the marker, so widening the marker no longer shifts the description off column.

For the confirm card the same join gives a label column that does not need recounting.

```go
label := lipgloss.NewStyle().Width(13)
row := lipgloss.JoinHorizontal(lipgloss.Top,
	label.Render("Dependencies"),
	": ",
	"Django, "+depsText)
```

### Where lipgloss earns its keep in this project

**Centering.** `Model` has no width field and `Update` handles `KeyPressMsg` and `PasteMsg` today. Add `case tea.WindowSizeMsg: m.width = msg.Width`, then hand the finished block to `lipgloss.PlaceHorizontal(m.width, lipgloss.Center, b.String())`. `PlaceHorizontal` pads each line in place and inserts no newlines, so the row counts survive untouched. `lipgloss.Place` also takes a height, and that path calls `PlaceVertical`, which prepends blank lines when it centers vertically. Every cursor row would then need that offset added. Skip vertical centering or pay for it in the count.

**Long names on narrow terminals.** `lipgloss.Width` tells you a line is too wide but does not shorten it. Cap the style with `unselectedItemStyle.MaxWidth(m.width - 8)` or truncate the text first. Note the shape of the v2 API here, `MaxWidth` is a method on `Style` and there is no package level `Truncate` in lipgloss v2. Rendering with `MaxWidth` truncates each line after borders and margins are applied. For a prefix such as `…`, use `ansi.Truncate` from `github.com/charmbracelet/x/ansi`, which `go.mod` already pins at v0.11.8.

**The description column.** Feed `m.width` into `subtitleStyle.MaxWidth(m.width - 8)` so descriptions wrap or clip instead of running off the edge. Do the same for the card with `cardStyle.Width(m.width - 4)`. The card currently has no width, so a long `depsText` line spills past the terminal edge, the terminal hard wraps it, and the rounded border characters land on the wrong rows. Wrapping the card is safe for the cursor math because only `nameRow` and `depRow` are cursor rows and neither comes out of a card.

### Where lipgloss does not help

`View` still assembles the frame top to bottom. lipgloss hands you functions over strings, not a widget tree, and it does not diff or schedule frames.

`JoinHorizontal` and `JoinVertical` allocate. Each one splits the blocks into line slices, pads them, and rebuilds one string. Nine dependency rows and a five line card is a handful of small allocations per frame. That is irrelevant at this scale.

`JoinVertical` writes a `"\n"` between lines and `PlaceHorizontal` pads without adding any, so `strings.Count(b.String(), "\n")` keeps working unchanged. `lipgloss.Height` counts rows the same way, by newlines.

The choice is about who measures width, not about performance.

### Recommendation for this repo

Convert now.

- Store `m.width` from `WindowSizeMsg` and center the frame with `PlaceHorizontal`. That is the one change that makes every other width decision possible, and it is the same work as item 1 above, so do them together.
- Convert the dependency row to `JoinHorizontal` with a computed gutter, as shown above. It removes the hand counted six spaces and gives you a measured indent.
- Cap names, descriptions, and the confirm card with `MaxWidth` or `Width` once `m.width` exists, and reach for `ansi.Truncate` where you want a `…` instead of a hard cut.

Leave alone.

- The `strings.Builder` assembly and the newline cursor counts. They are correct, they are easy to read, and a framework that replaces them costs more than it returns.
- The confirm card's label padding. The labels are static plain text and the spaces line up. Add `cardStyle.MaxWidth` to stop the card overflowing rather than rebuilding the rows. Only revisit the labels if you add one longer than `Dependencies`.
- `viewDocker`, `viewExecuting`, `viewDone`, and `viewError`. One line each, no width math, nothing to gain.

Use lipgloss wherever a width number meets styled text. Keep `fmt.Sprintf` where the text is plain.
