---
title: Process API
description: Complete reference for lightshell.process — scoped system command execution.
---

The `lightshell.process` module runs system commands from JavaScript. Commands are executed directly via Go's `exec.Command` — never through a shell — which prevents shell injection attacks. In restricted permission mode, commands must be explicitly whitelisted in `lightshell.json`. All methods are async and return Promises.

## Methods

### exec(cmd, args?, options?)

Execute a system command and return its output.

**Parameters:**
- `cmd` (string) — the command to execute (e.g., `"git"`, `"python3"`, `"ls"`)
- `args` (string[], optional) — array of command arguments (default: `[]`)
- `options` (object, optional):
  - `cwd` (string) — working directory for the command
  - `env` (object) — additional environment variables as key-value pairs
  - `timeout` (number) — timeout in milliseconds (default: no timeout)

**Returns:** `Promise<{ stdout: string, stderr: string, code: number }>` — the command result:
  - `stdout` — standard output as a string
  - `stderr` — standard error as a string
  - `code` — the exit code (`0` means success)

**Example:**
```js
// Simple command
const result = await lightshell.process.exec('git', ['status'])
console.log(result.stdout)
// "On branch main\nnothing to commit, working tree clean\n"

// Command with options
const result2 = await lightshell.process.exec('python3', ['script.py'], {
  cwd: '/path/to/project',
  env: { PYTHONPATH: '/custom/path' },
  timeout: 10000
})

if (result2.code !== 0) {
  console.error('Script failed:', result2.stderr)
}
```

**Errors:** Rejects if the command is not found, is not allowed by permissions, or the timeout expires. Note that a non-zero exit code does NOT cause a rejection — check `result.code` instead.

---

## Permission Scoping

In restricted permission mode, commands must be declared in `lightshell.json`. This lets you precisely control what a LightShell app can execute.

```json
{
  "permissions": {
    "process": {
      "exec": [
        { "cmd": "git", "args": ["status", "log", "diff"] },
        { "cmd": "python3", "args": ["*"] },
        { "cmd": "ls" }
      ]
    }
  }
}
```

| Config | Meaning |
|--------|---------|
| `{ "cmd": "git", "args": ["status", "log", "diff"] }` | Only `git status`, `git log`, and `git diff` are allowed |
| `{ "cmd": "python3", "args": ["*"] }` | `python3` with any arguments is allowed |
| `{ "cmd": "ls" }` | `ls` with no arguments or any arguments is allowed (omitting `args` means any) |

If no `permissions` key exists in `lightshell.json`, the app runs in permissive mode and all commands are allowed.

---

## Common Patterns

### Running Git Commands

```js
async function gitStatus(repoPath) {
  const result = await lightshell.process.exec('git', ['status', '--porcelain'], {
    cwd: repoPath
  })

  if (result.code !== 0) {
    throw new Error(`git error: ${result.stderr}`)
  }

  return result.stdout
    .split('\n')
    .filter(line => line.trim())
    .map(line => ({
      status: line.substring(0, 2).trim(),
      file: line.substring(3)
    }))
}

async function gitLog(repoPath, count = 10) {
  const result = await lightshell.process.exec(
    'git',
    ['log', `--oneline`, `-${count}`],
    { cwd: repoPath }
  )
  return result.stdout.trim().split('\n')
}
```

### Running Scripts

```js
async function runPythonScript(scriptPath, args = []) {
  const result = await lightshell.process.exec(
    'python3',
    [scriptPath, ...args],
    { timeout: 30000 }
  )

  if (result.code !== 0) {
    throw new Error(`Python error (exit ${result.code}):\n${result.stderr}`)
  }

  return result.stdout
}

// Usage
const output = await runPythonScript('/path/to/analyze.py', ['--input', 'data.csv'])
console.log(output)
```

### Command Output in UI

```js
async function runCommand(cmd, args) {
  const outputEl = document.getElementById('terminal-output')
  outputEl.textContent = `$ ${cmd} ${args.join(' ')}\n`

  try {
    const result = await lightshell.process.exec(cmd, args, { timeout: 15000 })

    if (result.stdout) {
      outputEl.textContent += result.stdout
    }
    if (result.stderr) {
      outputEl.textContent += result.stderr
    }
    outputEl.textContent += `\n[exit code: ${result.code}]`
  } catch (err) {
    outputEl.textContent += `\nError: ${err.message}`
  }
}
```

### Check if a Tool is Installed

```js
async function isInstalled(cmd) {
  try {
    const result = await lightshell.process.exec('which', [cmd])
    return result.code === 0
  } catch {
    return false
  }
}

// Usage
if (await isInstalled('ffmpeg')) {
  console.log('ffmpeg is available')
} else {
  console.log('ffmpeg is not installed')
}
```

## Security Notes

- Commands are executed directly via `exec.Command`, **never** through a shell (`sh -c`). This means shell features like pipes (`|`), redirects (`>`), globbing (`*`), and variable expansion (`$VAR`) do not work and cannot be exploited.
- To chain commands, make multiple `exec()` calls from JavaScript.
- In restricted mode, attempting to run a command not listed in `permissions.process.exec` results in a permission error with an AI-friendly message explaining what was attempted and what is allowed.
- The `PATH` used to resolve commands is restricted to standard system paths (`/usr/bin`, `/usr/local/bin`, etc.), not the user's full shell `PATH`.

## Platform Notes

- On macOS, common tools like `git`, `python3`, and `open` are located in `/usr/bin/` or `/usr/local/bin/`.
- On Linux, tool locations vary by distribution but standard paths are searched.
- The `cwd` option sets the working directory for the child process only. It does not affect the LightShell app itself.
- The `env` option adds to (does not replace) the default environment variables. Use it to set variables like `LANG`, `PYTHONPATH`, or custom configuration.
- `timeout` causes the process to be killed (SIGKILL) and the Promise to reject if the command does not complete within the specified time.
