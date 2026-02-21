---
title: App Identifier
description: Choose and configure a unique app ID for your LightShell app.
---

Every LightShell app has an identifier — a reverse-domain string like `com.example.myapp` that uniquely identifies your application. This identifier is used across both platforms for bundle identification, data storage paths, desktop integration, and code signing. Choose it carefully before your first release, because changing it later has consequences.

## Setting the App ID

Configure the app ID in `lightshell.json`:

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "build": {
    "appId": "com.example.myapp"
  }
}
```

## Naming Convention

App IDs follow the reverse-domain naming convention:

```
{tld}.{domain}.{appname}
```

Examples:

| App ID | Explanation |
|--------|-------------|
| `com.yourname.todoapp` | Personal project using your name |
| `com.yourcompany.dashboard` | Company app |
| `dev.lightshell.myapp` | Using a `.dev` domain |
| `io.github.yourname.myapp` | GitHub-based project |
| `org.example.editor` | Organization project |

## Naming Rules

Follow these rules when choosing an app ID:

- **Use lowercase letters only.** Uppercase is technically valid but discouraged and can cause issues on case-sensitive Linux filesystems.
- **Use only alphanumeric characters and dots.** Hyphens are allowed in some contexts but dots are the standard separator. Avoid underscores, spaces, and special characters.
- **Start with a domain you own or control.** This prevents collisions with other developers' apps. If you do not own a domain, use `io.github.{username}` or similar.
- **End with a descriptive app name.** This should identify the specific application.
- **Keep it reasonably short.** Long identifiers work but are cumbersome in logs and file paths.
- **Make it globally unique.** No two apps should share the same identifier. The reverse-domain convention handles this naturally if you use a domain you control.

## Default Value

If you do not set `build.appId` in your `lightshell.json`, LightShell generates a default:

```
com.lightshell.{name}
```

Where `{name}` is the `name` field from your config. For example, if your app is named `my-app`, the default app ID is `com.lightshell.my-app`.

This default works for development, but you should set a proper app ID before distributing your app.

## Where the App ID Is Used

The app ID appears in several places across both platforms:

### macOS Bundle Identifier

The app ID becomes the `CFBundleIdentifier` in your `.app` bundle's `Info.plist`:

```xml
<key>CFBundleIdentifier</key>
<string>com.example.myapp</string>
```

macOS uses the bundle identifier to:

- Track app preferences in `NSUserDefaults`
- Associate documents with your app
- Identify your app for code signing and notarization
- Manage Gatekeeper and security settings
- Store per-app permissions (camera, microphone, location)

### Linux Desktop Entry

The app ID is used in the `.desktop` file that registers your app with the system:

```ini
[Desktop Entry]
Type=Application
Name=My App
Exec=/usr/bin/myapp
Icon=com.example.myapp
Categories=Utility;
```

Desktop environments use this to identify your app in the launcher, taskbar, and notification system.

### Data Directory

The app ID determines where your app stores persistent data:

| Platform | Data Directory |
|----------|---------------|
| macOS | `~/Library/Application Support/{appId}` |
| Linux | `~/.local/share/{appId}` |

For example, with `com.example.myapp`:

- macOS: `~/Library/Application Support/com.example.myapp/`
- Linux: `~/.local/share/com.example.myapp/`

The `lightshell.store` API and `lightshell.app.dataDir()` both use this path. All persistent data — the bbolt database, saved files, caches — lives here.

### Code Signing

On macOS, the app ID is embedded in the code signature. Apple's notarization service records it. Changing the app ID after notarization means re-signing and re-notarizing.

### Protocol Handler

If you register a custom URL protocol (deep links), the app ID is used to associate the protocol with your app at the OS level.

## Changing the App ID

Changing the app ID after release has real consequences:

**Data directory changes.** The persistent data directory path includes the app ID. If you change the app ID from `com.example.myapp` to `com.example.newname`, the data directory changes from `~/Library/Application Support/com.example.myapp/` to `~/Library/Application Support/com.example.newname/`. Existing user data is orphaned in the old directory — the app will not find it.

**macOS identity changes.** macOS treats a new bundle identifier as a completely different application. Users lose:

- Dock position
- Login item status
- Notification preferences
- Accessibility permissions
- Any other per-app OS settings

**Code signing breaks.** If you have signed or notarized with the old identifier, you need to re-sign and re-notarize with the new one.

If you must change the app ID after release, consider adding migration logic to your app that checks for the old data directory and copies data to the new one:

```js
async function migrateData() {
  const dataDir = await lightshell.app.dataDir()
  const platform = await lightshell.system.platform()

  let oldDir
  if (platform === 'darwin') {
    oldDir = (await lightshell.system.homeDir()) +
      '/Library/Application Support/com.example.oldname'
  } else {
    oldDir = (await lightshell.system.homeDir()) +
      '/.local/share/com.example.oldname'
  }

  if (await lightshell.fs.exists(oldDir)) {
    // Copy data from old location to new
    const oldData = await lightshell.fs.readFile(oldDir + '/store.db')
    await lightshell.fs.mkdir(dataDir)
    await lightshell.fs.writeFile(dataDir + '/store.db', oldData)
  }
}
```

## Best Practices

**Set it early.** Choose your app ID before writing any persistence code or distributing any builds.

**Use a domain you control.** This guarantees uniqueness without coordination. If you later sell or transfer the app, the identifier remains stable.

**Do not include version numbers.** The app ID should be stable across all versions of your app. Version information belongs in the `version` field.

**Avoid generic names.** `com.example.app` is likely to collide. Use a specific, descriptive name for the app portion.

**Keep it consistent with your other apps.** If you ship multiple apps, use the same domain prefix:

```
com.yourcompany.editor
com.yourcompany.viewer
com.yourcompany.converter
```

This makes it easy to identify all your apps in the filesystem and in system preferences.
