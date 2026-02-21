---
title: "5. Adding Persistence"
description: Store and retrieve data across app restarts using lightshell.store.
---

Most apps need to remember things between sessions. User preferences, saved data, recent files — all of this requires persistence. LightShell provides `lightshell.store`, a key-value storage API backed by bbolt, so you can save and load data without touching the file system directly.

## Why Not Just Use Files?

You could use `lightshell.fs.readFile` and `lightshell.fs.writeFile` to persist data. It works, but it is tedious:

```js
// The hard way — manual file-based persistence
const dataDir = await lightshell.app.dataDir()
const path = dataDir + '/todos.json'

// Save
await lightshell.fs.writeFile(path, JSON.stringify(todos))

// Load
const raw = await lightshell.fs.readFile(path)
const todos = JSON.parse(raw)
```

You have to manage file paths, serialize to JSON, handle missing files on first launch, and deal with write errors. `lightshell.store` eliminates all of that.

## The Store API

`lightshell.store` is a persistent key-value store. Values are automatically serialized and deserialized — you pass in objects and get objects back. The data is saved to a database file in your app's data directory and survives restarts.

```js
// Save a value
await lightshell.store.set('username', 'Alice')

// Read it back
const name = await lightshell.store.get('username')
console.log(name) // "Alice"

// Works with objects, arrays, numbers — anything JSON-serializable
await lightshell.store.set('settings', { theme: 'dark', fontSize: 14 })
const settings = await lightshell.store.get('settings')
console.log(settings.theme) // "dark"
```

No file paths. No `JSON.stringify`. Just `get` and `set`.

## Building a Persistent Todo List

Let's build a todo list that remembers your items across app restarts. Start with the HTML structure in `src/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Todo List</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <h1>Todos</h1>
  <div class="input-row">
    <input type="text" id="todo-input" placeholder="What needs to be done?">
    <button id="add-btn">Add</button>
  </div>
  <ul id="todo-list"></ul>
  <div class="footer">
    <span id="count">0 items</span>
    <button id="clear-btn">Clear All</button>
  </div>
  <script src="app.js"></script>
</body>
</html>
```

### Loading Todos on Startup

In `src/app.js`, start by loading any saved todos when the app launches:

```js
let todos = []

async function loadTodos() {
  const saved = await lightshell.store.get('todos')
  if (saved) {
    todos = saved
  }
  render()
}
```

`lightshell.store.get` returns `null` if the key does not exist, so on first launch `saved` will be `null` and we start with an empty array.

### Saving Todos

Every time the list changes, save it:

```js
async function saveTodos() {
  await lightshell.store.set('todos', todos)
}
```

That is the entire persistence layer. One line.

### Adding and Deleting Items

Wire up the input and buttons:

```js
async function addTodo() {
  const input = document.getElementById('todo-input')
  const text = input.value.trim()
  if (!text) return

  todos.push({ text, done: false })
  input.value = ''
  await saveTodos()
  render()
}

async function deleteTodo(index) {
  todos.splice(index, 1)
  await saveTodos()
  render()
}

async function toggleTodo(index) {
  todos[index].done = !todos[index].done
  await saveTodos()
  render()
}
```

### Rendering the List

```js
function render() {
  const list = document.getElementById('todo-list')
  list.innerHTML = todos.map((todo, i) => `
    <li class="${todo.done ? 'done' : ''}">
      <input type="checkbox" ${todo.done ? 'checked' : ''}
             onchange="toggleTodo(${i})">
      <span>${todo.text}</span>
      <button onclick="deleteTodo(${i})">Delete</button>
    </li>
  `).join('')

  document.getElementById('count').textContent =
    `${todos.filter(t => !t.done).length} items remaining`
}
```

### Clearing All Todos

The clear button deletes the stored data entirely:

```js
async function clearAll() {
  const confirmed = await lightshell.dialog.confirm(
    'Clear Todos',
    'Are you sure you want to delete all todos?'
  )
  if (!confirmed) return

  todos = []
  await lightshell.store.delete('todos')
  render()
}
```

Here we use `lightshell.store.delete` to remove the key entirely rather than saving an empty array. Both approaches work — `delete` is cleaner when you want to reset to a "never been set" state.

### Wiring It Up

Connect the event listeners and load on startup:

```js
document.getElementById('add-btn').addEventListener('click', addTodo)
document.getElementById('clear-btn').addEventListener('click', clearAll)
document.getElementById('todo-input').addEventListener('keydown', (e) => {
  if (e.key === 'Enter') addTodo()
})

loadTodos()
```

Run `lightshell dev`, add a few todos, then quit and relaunch. Your todos are still there.

## Other Store Methods

### Checking If a Key Exists

Use `has` to check for existence without reading the value:

```js
const hasTodos = await lightshell.store.has('todos')
if (hasTodos) {
  console.log('Found saved todos')
}
```

### Listing Keys

Use `keys` to list all stored keys, optionally filtered by prefix:

```js
// Store some settings
await lightshell.store.set('settings.theme', 'dark')
await lightshell.store.set('settings.fontSize', 14)
await lightshell.store.set('user.name', 'Alice')

// List all keys starting with "settings."
const settingKeys = await lightshell.store.keys('settings.*')
console.log(settingKeys) // ["settings.theme", "settings.fontSize"]

// List all keys
const allKeys = await lightshell.store.keys()
console.log(allKeys) // ["settings.theme", "settings.fontSize", "user.name"]
```

### Clearing Everything

`clear` removes all stored data. Use with caution:

```js
await lightshell.store.clear()
```

## Common Patterns

### Persisting App Settings

```js
const defaults = { theme: 'light', fontSize: 14, sidebarOpen: true }

async function loadSettings() {
  const saved = await lightshell.store.get('settings')
  return { ...defaults, ...saved }
}

async function saveSetting(key, value) {
  const settings = await loadSettings()
  settings[key] = value
  await lightshell.store.set('settings', settings)
}
```

### Recently Opened Files

```js
async function addRecentFile(path) {
  let recent = (await lightshell.store.get('recentFiles')) || []
  recent = recent.filter(f => f !== path)  // remove duplicates
  recent.unshift(path)                      // add to front
  recent = recent.slice(0, 10)              // keep last 10
  await lightshell.store.set('recentFiles', recent)
}
```

## Recap

You have learned how to:

1. **Save data** with `lightshell.store.set` — no file paths or serialization needed
2. **Load data** with `lightshell.store.get` — returns `null` if the key does not exist
3. **Delete data** with `lightshell.store.delete` and `lightshell.store.clear`
4. **Check and list** keys with `lightshell.store.has` and `lightshell.store.keys`
5. **Build persistent apps** that remember state across restarts

The store handles all the complexity of file paths, serialization, and atomicity behind a simple key-value interface.

Next, let's connect to external APIs.
