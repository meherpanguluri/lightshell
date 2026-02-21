---
title: Store API
description: Complete reference for lightshell.store — persistent key-value storage.
---

The `lightshell.store` module provides persistent key-value storage backed by [bbolt](https://github.com/etcd-io/bbolt). Values are automatically serialized to and from JSON, so you can store strings, numbers, booleans, arrays, and objects without manual serialization. Data persists across app restarts. All methods are async and return Promises.

The store database file is located at `$APP_DATA/store.db` and is created automatically on first use.

## Methods

### get(key)

Get a value by key. Returns `null` if the key does not exist.

**Parameters:**
- `key` (string) — the key to look up

**Returns:** `Promise<any | null>` — the stored value (automatically deserialized from JSON), or `null` if not found

**Example:**
```js
const name = await lightshell.store.get('user.name')
console.log(name) // "Alice" or null

const settings = await lightshell.store.get('settings')
console.log(settings) // { theme: "dark", fontSize: 14 } or null
```

---

### set(key, value)

Set a key-value pair. If the key already exists, its value is replaced. Values are automatically JSON-serialized.

**Parameters:**
- `key` (string) — the key to set
- `value` (any) — the value to store (must be JSON-serializable)

**Returns:** `Promise<void>`

**Example:**
```js
// Store a string
await lightshell.store.set('user.name', 'Alice')

// Store a number
await lightshell.store.set('app.launchCount', 42)

// Store an object
await lightshell.store.set('settings', {
  theme: 'dark',
  fontSize: 14,
  sidebarWidth: 250
})

// Store an array
await lightshell.store.set('recentFiles', [
  '/Users/alice/notes.md',
  '/Users/alice/todo.txt'
])
```

---

### delete(key)

Delete a key-value pair. No error is thrown if the key does not exist.

**Parameters:**
- `key` (string) — the key to delete

**Returns:** `Promise<void>`

**Example:**
```js
await lightshell.store.delete('user.name')
```

---

### has(key)

Check whether a key exists in the store.

**Parameters:**
- `key` (string) — the key to check

**Returns:** `Promise<boolean>` — `true` if the key exists, `false` otherwise

**Example:**
```js
const exists = await lightshell.store.has('user.name')
if (!exists) {
  await lightshell.store.set('user.name', 'Anonymous')
}
```

---

### keys(prefix?)

List all keys in the store, optionally filtered by a prefix. Call with no arguments or an empty string to list all keys.

**Parameters:**
- `prefix` (string, optional) — only return keys that start with this prefix

**Returns:** `Promise<string[]>` — array of matching key names

**Example:**
```js
// List all keys
const allKeys = await lightshell.store.keys()
console.log(allKeys) // ["user.name", "user.email", "settings", "recentFiles"]

// List keys with prefix
const userKeys = await lightshell.store.keys('user.')
console.log(userKeys) // ["user.name", "user.email"]
```

---

### clear()

Delete all key-value pairs from the store. This is irreversible.

**Parameters:** none

**Returns:** `Promise<void>`

**Example:**
```js
const confirmed = await lightshell.dialog.confirm(
  'Clear Data',
  'This will delete all saved data. Are you sure?'
)
if (confirmed) {
  await lightshell.store.clear()
}
```

---

## Common Patterns

### Settings Management

```js
const DEFAULT_SETTINGS = {
  theme: 'light',
  fontSize: 14,
  autoSave: true,
  sidebarVisible: true
}

async function loadSettings() {
  const saved = await lightshell.store.get('settings')
  return { ...DEFAULT_SETTINGS, ...saved }
}

async function saveSetting(key, value) {
  const settings = await loadSettings()
  settings[key] = value
  await lightshell.store.set('settings', settings)
  return settings
}

// Usage
const settings = await loadSettings()
applyTheme(settings.theme)

// Update a single setting
await saveSetting('theme', 'dark')
```

### Todo List Persistence

```js
async function loadTodos() {
  return (await lightshell.store.get('todos')) || []
}

async function addTodo(text) {
  const todos = await loadTodos()
  todos.push({ id: Date.now(), text, done: false })
  await lightshell.store.set('todos', todos)
  return todos
}

async function toggleTodo(id) {
  const todos = await loadTodos()
  const todo = todos.find(t => t.id === id)
  if (todo) {
    todo.done = !todo.done
    await lightshell.store.set('todos', todos)
  }
  return todos
}

async function deleteTodo(id) {
  const todos = await loadTodos()
  const filtered = todos.filter(t => t.id !== id)
  await lightshell.store.set('todos', filtered)
  return filtered
}
```

### Caching API Responses

```js
async function cachedFetch(url, ttlMs = 60000) {
  const cacheKey = `cache.${url}`
  const cached = await lightshell.store.get(cacheKey)

  if (cached && Date.now() - cached.timestamp < ttlMs) {
    return cached.data
  }

  const response = await lightshell.http.fetch(url)
  const data = JSON.parse(response.body)

  await lightshell.store.set(cacheKey, {
    data,
    timestamp: Date.now()
  })

  return data
}

// Usage — cached for 5 minutes
const weather = await cachedFetch('https://api.weather.com/current', 300000)
```

## Platform Notes

- The store database is a single file at `$APP_DATA/store.db` (see `lightshell.app.dataDir()` for the exact location)
- On macOS, `$APP_DATA` resolves to `~/Library/Application Support/{appId}/`
- On Linux, `$APP_DATA` resolves to `~/.config/{appId}/`
- The database file is created automatically on the first `set()` call
- Values must be JSON-serializable. Functions, `undefined`, circular references, and other non-serializable values will cause an error.
- Keys are plain strings. Dot notation (e.g., `user.name`) is a naming convention, not structural — the store does not interpret dots as hierarchy.
- The store is backed by bbolt, a pure-Go embedded key-value database. It is safe for concurrent reads and serializes writes automatically.
- There is no size limit on individual values, but storing very large values (>10MB) may impact performance.
