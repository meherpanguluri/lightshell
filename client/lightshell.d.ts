declare global {
  interface Window {
    lightshell: LightShell
  }
}

interface LightShell {
  window: LightShellWindow
  fs: LightShellFS
  dialog: LightShellDialog
  clipboard: LightShellClipboard
  shell: LightShellShell
  notify: LightShellNotify
  tray: LightShellTray
  menu: LightShellMenu
  system: LightShellSystem
  app: LightShellApp
  on(event: string, callback: (data: any) => void): () => void
}

interface LightShellWindow {
  setTitle(title: string): Promise<void>
  setSize(width: number, height: number): Promise<void>
  getSize(): Promise<{ width: number; height: number }>
  setPosition(x: number, y: number): Promise<void>
  getPosition(): Promise<{ x: number; y: number }>
  minimize(): Promise<void>
  maximize(): Promise<void>
  fullscreen(): Promise<void>
  restore(): Promise<void>
  close(): Promise<void>
  onResize(callback: (data: { width: number; height: number }) => void): () => void
  onMove(callback: (data: { x: number; y: number }) => void): () => void
  onFocus(callback: () => void): () => void
  onBlur(callback: () => void): () => void
}

interface LightShellFS {
  readFile(path: string, encoding?: string): Promise<string>
  writeFile(path: string, data: string): Promise<void>
  readDir(path: string): Promise<FileEntry[]>
  exists(path: string): Promise<boolean>
  stat(path: string): Promise<FileStat>
  mkdir(path: string): Promise<void>
  remove(path: string): Promise<void>
  watch(path: string, callback: (event: FileWatchEvent) => void): () => void
}

interface FileEntry {
  name: string
  isDir: boolean
  size: number
}

interface FileStat {
  name: string
  size: number
  isDir: boolean
  modTime: string
  mode: string
}

interface FileWatchEvent {
  path: string
  op: 'create' | 'write' | 'remove' | 'rename'
}

interface DialogOpenOptions {
  title?: string
  filters?: Array<{ name: string; extensions: string[] }>
  multiple?: boolean
  directory?: boolean
  defaultPath?: string
}

interface DialogSaveOptions {
  title?: string
  filters?: Array<{ name: string; extensions: string[] }>
  defaultPath?: string
}

interface LightShellDialog {
  open(options?: DialogOpenOptions): Promise<string | string[] | null>
  save(options?: DialogSaveOptions): Promise<string | null>
  message(title: string, message: string): Promise<void>
  confirm(title: string, message: string): Promise<boolean>
  prompt(title: string, defaultValue?: string): Promise<string | null>
}

interface LightShellClipboard {
  read(): Promise<string>
  write(text: string): Promise<void>
}

interface LightShellShell {
  open(url: string): Promise<void>
}

interface NotifyOptions {
  icon?: string
}

interface LightShellNotify {
  send(title: string, body: string, options?: NotifyOptions): Promise<void>
}

interface TrayOptions {
  icon?: string
  tooltip?: string
  menu?: MenuTemplate[]
}

interface LightShellTray {
  set(options: TrayOptions): Promise<void>
  remove(): Promise<void>
  onClick(callback: () => void): () => void
}

interface MenuTemplate {
  label: string
  click?: string
  role?: string
  submenu?: MenuTemplate[]
  separator?: boolean
}

interface LightShellMenu {
  set(template: MenuTemplate[]): Promise<void>
}

interface LightShellSystem {
  platform(): Promise<'darwin' | 'linux'>
  arch(): Promise<string>
  homeDir(): Promise<string>
  tempDir(): Promise<string>
  hostname(): Promise<string>
}

interface LightShellApp {
  quit(): Promise<void>
  version(): Promise<string>
  dataDir(): Promise<string>
}

export {}
