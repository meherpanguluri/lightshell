---
title: Example Prompts
description: Copy-paste prompts for AI to build LightShell apps.
---

Each prompt below is a complete, self-contained instruction that you can paste into Claude, ChatGPT, Cursor, or any other AI tool. Each one produces a working LightShell app. Adjust the details to match your needs.

For best results, also include the LightShell API reference as context. Prepend this line to any prompt:

```
Use this API reference: https://lightshell.dev/llms-full.txt
```

---

## 1. Todo App with Persistence

**Uses:** `lightshell.store`, `lightshell.dialog.confirm`, `lightshell.window.setTitle`

```
Build a LightShell desktop app: a todo list manager.

Requirements:
- A text input at the top to add new todos, with an "Add" button
- Each todo has a checkbox to mark complete and a delete button
- Todos persist across app restarts using lightshell.store
- Store todos under the key 'todos' as a JSON array of {id, text, done, createdAt}
- Show a count of remaining items: "3 items left"
- Filter buttons: All, Active, Completed
- "Clear completed" button that asks for confirmation with lightshell.dialog.confirm
- Update the window title with the count: "Todos (3)" via lightshell.window.setTitle
- Clean, minimal design with a white background and subtle borders
- Keyboard: Enter in the input field adds the todo

Put all code in three files: src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "todo-app", version "1.0.0", window 500x700.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 2. Markdown Note Editor

**Uses:** `lightshell.fs`, `lightshell.dialog.open`, `lightshell.dialog.save`, `lightshell.window.setTitle`, `lightshell.store`

```
Build a LightShell desktop app: a Markdown note editor.

Requirements:
- Split-pane layout: editor on the left, live preview on the right
- The editor is a <textarea> with a monospace font
- The preview renders Markdown to HTML (use marked.js from CDN:
  https://cdn.jsdelivr.net/npm/marked/marked.min.js)
- File menu operations via buttons in a toolbar at the top:
  - "New" — clears the editor (confirm if unsaved changes)
  - "Open" — uses lightshell.dialog.open() to pick a .md file, then loads it with lightshell.fs.readFile
  - "Save" — if a file is already open, saves to that path. If not, uses lightshell.dialog.save()
  - "Save As" — always shows lightshell.dialog.save()
- Track whether the document has unsaved changes (dirty flag)
- Window title shows the filename and a bullet if dirty: "notes.md *" via lightshell.window.setTitle
- Remember the last opened file path using lightshell.store.set('lastFile', path)
- On launch, check lightshell.store for lastFile and offer to reopen it
- Keyboard: Cmd/Ctrl+S to save, Cmd/Ctrl+O to open
- The preview should update as you type (debounced, 300ms delay)

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "md-editor", version "1.0.0", window 1000x700.
All lightshell.* calls are async — use await. No Node.js APIs.
```

---

## 3. API Dashboard

**Uses:** `lightshell.http.fetch`, `lightshell.store`, `lightshell.dialog.prompt`

```
Build a LightShell desktop app: an API endpoint monitoring dashboard.

Requirements:
- Users can add API endpoints to monitor (URL + name + expected status code)
- Use lightshell.dialog.prompt to ask for the endpoint URL and name when adding
- All endpoints stored in lightshell.store under key 'endpoints'
- A "Check All" button that pings every endpoint using lightshell.http.fetch and records:
  - HTTP status code
  - Response time (ms)
  - Whether status matches expected
- Display results in a card grid:
  - Green card = status matches expected
  - Red card = status mismatch or error
  - Each card shows: name, URL, status, response time, last checked time
- Auto-refresh every 60 seconds (configurable with a dropdown: 30s, 60s, 5min, off)
- "Delete endpoint" button on each card
- A summary bar at the top: "5/6 endpoints healthy"
- Clean dashboard aesthetic with a dark sidebar listing endpoint names

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "api-dashboard", version "1.0.0", window 1100x750.
Use lightshell.http.fetch (not browser fetch) for CORS-free requests.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 4. Clipboard Manager

**Uses:** `lightshell.clipboard`, `lightshell.shortcuts`, `lightshell.store`, `lightshell.notify`

```
Build a LightShell desktop app: a clipboard history manager.

Requirements:
- Poll lightshell.clipboard.read() every 2 seconds to detect new clipboard content
- Store up to 50 clipboard entries in lightshell.store under key 'history'
  Each entry: {text, timestamp, pinned}
- Display the history as a scrollable list, newest first
- Clicking an entry copies it back to the clipboard with lightshell.clipboard.write()
  and shows a notification via lightshell.notify.send({title: "Copied", body: text.slice(0, 50)})
- "Pin" button on each entry to keep it at the top permanently
- "Delete" button to remove individual entries
- "Clear All" button (does not remove pinned entries)
- Search bar at the top to filter entries by text content
- Register a global shortcut: CommandOrControl+Shift+V to bring the window to focus
  using lightshell.shortcuts.register
- Compact UI: each entry shows the first 100 characters with a timestamp below it
- Duplicate detection: do not add an entry if the text matches the most recent one

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "clipboard-manager", version "1.0.0", window 400x600.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 5. File Explorer

**Uses:** `lightshell.fs`, `lightshell.dialog`, `lightshell.shell.open`, `lightshell.system`

```
Build a LightShell desktop app: a simple file explorer.

Requirements:
- On launch, show the user's home directory (get it with lightshell.system.homeDir())
- Display directory contents using lightshell.fs.readDir, which returns [{name, isDir}]
- Show folders first (with a folder icon), then files (with a file icon)
- Use Unicode characters for icons: folders = folder icon, files = page icon
- Clicking a folder navigates into it
- Clicking a file opens it with lightshell.shell.open(fullPath)
- Breadcrumb navigation at the top showing the current path as clickable segments
- "Up" button to go to the parent directory
- Show file size and last modified date from lightshell.fs.stat for each item
- "Open folder..." button that uses lightshell.dialog.open() with directory mode to jump anywhere
- Right-click context menu is not needed — keep it simple
- Path bar at the top shows the full current path
- Status bar at the bottom: "12 items (8 files, 4 folders)"
- Sort by name, size, or date (toggle buttons)
- Handle permission errors gracefully with try/catch — show "Access denied" for
  unreadable directories

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "file-explorer", version "1.0.0", window 900x650.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 6. Color Picker Tool

**Uses:** `lightshell.clipboard`, `lightshell.notify`, `lightshell.store`

```
Build a LightShell desktop app: a color picker and palette manager.
Put everything in a single file: src/index.html (inline CSS and JS).

Requirements:
- A large color picker using <input type="color">
- Display the selected color in HEX, RGB, and HSL formats
- A "Copy" button next to each format that copies to clipboard with
  lightshell.clipboard.write(value) and shows a notification
- A palette section below where users can save colors
- "Add to palette" button saves the current color to lightshell.store
  under key 'palette' as an array of {hex, name} objects
- Each palette swatch is a 40x40 colored square that, when clicked, loads that
  color into the picker
- "Rename" each palette entry by clicking its label (use lightshell.dialog.prompt)
- "Delete" button on each swatch
- A "lighter" and "darker" row showing 5 tint/shade variations of the current color
- The app should be compact — designed for a small window
- Show a contrast checker: display the current color as background with white and
  black text, showing WCAG contrast ratios for each

Also generate lightshell.json with name "color-picker", version "1.0.0", window 450x650.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 7. System Monitor

**Uses:** `lightshell.system`, `lightshell.process.exec`

```
Build a LightShell desktop app: a system information and monitor dashboard.

Requirements:
- Top section: system info from lightshell.system
  - Platform (lightshell.system.platform()), Architecture (lightshell.system.arch())
  - Hostname (lightshell.system.hostname())
  - Home directory (lightshell.system.homeDir())
- Middle section: live stats refreshed every 3 seconds using lightshell.process.exec:
  - CPU usage: run 'top -l 1 -n 0' on macOS or 'cat /proc/stat' on Linux, parse the output
  - Memory: run 'vm_stat' on macOS or 'free -m' on Linux
  - Disk: run 'df -h /' to get disk usage
  - Display these as simple progress bars with percentage labels
- Bottom section: running processes
  - Run 'ps aux --sort=-%mem' (Linux) or 'ps aux -m' (macOS) and show top 10 by memory
  - Display in a table: PID, Name, CPU%, MEM%, Command
- Detect the platform with lightshell.system.platform() and run the appropriate commands
- A "Refresh" button to manually refresh all stats
- Auto-refresh toggle (on by default, every 3 seconds)
- Clean, dashboard-style layout with dark background and colored accent bars

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "system-monitor", version "1.0.0", window 800x700.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 8. RSS Reader

**Uses:** `lightshell.http.fetch`, `lightshell.store`, `lightshell.shell.open`, `lightshell.dialog.prompt`

```
Build a LightShell desktop app: an RSS feed reader.

Requirements:
- Sidebar on the left lists saved RSS feeds by name
- "Add Feed" button at the bottom of the sidebar, uses lightshell.dialog.prompt to
  ask for the feed URL, then fetches it with lightshell.http.fetch to get the title
- Feeds stored in lightshell.store under key 'feeds' as [{url, title}]
- Clicking a feed in the sidebar fetches the RSS XML with lightshell.http.fetch,
  parses it with DOMParser, and displays the entries in the main content area
- Each entry shows: title, published date, and a short excerpt (first 200 chars of description)
- Clicking an entry title opens the link in the system browser via lightshell.shell.open
- "Remove Feed" option on each feed in the sidebar
- "Refresh" button to re-fetch the current feed
- "Mark all read" button — track read state in lightshell.store under 'readItems' as a Set of URLs
- Unread items shown in bold, read items in normal weight
- Unread count shown next to each feed name in the sidebar: "Hacker News (12)"
- Three-column layout would be too complex — use two columns: sidebar + main content
- Handle fetch errors gracefully — show "Could not load feed" with a retry button

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "rss-reader", version "1.0.0", window 1000x700.
Use lightshell.http.fetch (not browser fetch) — RSS feeds are cross-origin.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 9. Bookmark Manager

**Uses:** `lightshell.store`, `lightshell.http.fetch`, `lightshell.shell.open`, `lightshell.dialog`

```
Build a LightShell desktop app: a bookmark manager.

Requirements:
- Add bookmarks by pasting a URL — the app fetches the page title automatically
  using lightshell.http.fetch (parse <title> from the HTML)
- Each bookmark has: url, title, tags (array), createdAt
- All bookmarks stored in lightshell.store under key 'bookmarks'
- Tag system: assign tags when adding a bookmark using lightshell.dialog.prompt
  (comma-separated input). Display tags as colored pills.
- Sidebar shows all unique tags with counts. Clicking a tag filters the bookmark list.
- "All Bookmarks" option at the top of the sidebar
- Search bar filters bookmarks by title or URL as you type
- Clicking a bookmark opens it in the system browser via lightshell.shell.open
- Edit and delete buttons on each bookmark
- Export feature: "Export as HTML" generates a bookmarks.html file and lets the user
  choose the save path with lightshell.dialog.save(), then writes with lightshell.fs.writeFile
- Import feature: lightshell.dialog.open() to select an HTML bookmarks file,
  parse it, and merge into existing bookmarks
- Sort by: date added, title, or most recently clicked

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "bookmark-manager", version "1.0.0", window 950x700.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## 10. Pomodoro Timer

**Uses:** `lightshell.notify`, `lightshell.store`, `lightshell.window.setTitle`, `lightshell.shortcuts`

```
Build a LightShell desktop app: a Pomodoro timer with session tracking.

Requirements:
- Large circular timer display showing minutes and seconds remaining
- Three modes: Focus (25 min), Short Break (5 min), Long Break (15 min)
- Start / Pause / Reset buttons
- When a session completes, send a system notification via
  lightshell.notify.send({title: "Pomodoro", body: "Focus session complete! Take a break."})
- Auto-advance: after Focus, switch to Short Break. After 4 Focus sessions, switch to Long Break.
- Update window title with remaining time: "Focus 23:45" via lightshell.window.setTitle
- Session history: track completed sessions in lightshell.store under 'sessions'
  as [{type, completedAt, duration}]
- "Today" stats section below the timer: focus sessions completed, total focus time
- "History" view showing the last 7 days as a simple bar chart (CSS-only, no chart library)
- Settings: customize focus/break durations, stored in lightshell.store under 'settings'
- Global shortcut: CommandOrControl+Shift+Space to start/pause the timer
  via lightshell.shortcuts.register
- Visually distinct modes: warm red for focus, cool green for short break, blue for long break

Put all code in src/index.html, src/app.js, src/style.css.
Also generate lightshell.json with name "pomodoro", version "1.0.0", window 400x650.
All lightshell.* calls are async — use await. No Node.js APIs, no npm.
```

---

## Tips for Customizing These Prompts

- **Change the window size** — adjust the dimensions in the lightshell.json line to match your desired layout
- **Add more features** — append additional bullet points to the requirements list
- **Change the styling** — add a line like "Use a dark theme with blue accents" or "Match the macOS system look"
- **Combine apps** — take features from multiple prompts to build a more complex tool
- **Add error handling** — append "Handle all errors with try/catch and show lightshell.dialog.message on failure"
