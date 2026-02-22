// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeBlack from 'starlight-theme-black';

// https://astro.build/config
export default defineConfig({
  site: 'https://lightshell.dev',
  base: '/docs',
  integrations: [
    starlight({
      title: 'LightShell Docs',
      tagline: 'Build native desktop apps with JS. Ship under 5MB.',
      plugins: [starlightThemeBlack({
        footerText: 'Built with [LightShell](https://lightshell.dev). Open source on [GitHub](https://github.com/lightshell-dev/lightshell).',
      })],
      customCss: ['./src/styles/custom.css'],
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/lightshell-dev/lightshell',
        },
      ],
      head: [
        {
          tag: 'link',
          attrs: {
            rel: 'preconnect',
            href: 'https://fonts.googleapis.com',
          },
        },
        {
          tag: 'link',
          attrs: {
            rel: 'preconnect',
            href: 'https://fonts.gstatic.com',
            crossorigin: '',
          },
        },
        {
          tag: 'link',
          attrs: {
            rel: 'stylesheet',
            href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap',
          },
        },
      ],
      expressiveCode: {
        themes: ['github-light', 'github-dark'],
        styleOverrides: {
          borderRadius: '12px',
          codeFontFamily: "'JetBrains Mono', ui-monospace, 'SF Mono', Menlo, Consolas, monospace",
          codeFontSize: '0.875rem',
        },
      },
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Quick Start', slug: 'getting-started' },
          ],
        },
        {
          label: 'Tutorial',
          items: [
            { label: '1. Your First App', slug: 'tutorial/01-your-first-app' },
            { label: '2. Native APIs', slug: 'tutorial/02-native-apis' },
            { label: '3. Styling', slug: 'tutorial/03-styling' },
            { label: '4. Packaging', slug: 'tutorial/04-packaging' },
            { label: '5. Adding Persistence', slug: 'tutorial/05-adding-persistence' },
            { label: '6. Connecting to APIs', slug: 'tutorial/06-connecting-to-apis' },
          ],
        },
        {
          label: 'Guides',
          items: [
            { label: 'File System', slug: 'guides/file-system' },
            { label: 'Menus & Tray', slug: 'guides/menus-tray' },
            { label: 'Dialogs & Prompts', slug: 'guides/dialogs-and-prompts' },
            { label: 'Keyboard Shortcuts', slug: 'guides/keyboard-shortcuts' },
            { label: 'Default Styles', slug: 'guides/default-css' },
            { label: 'Deep Linking', slug: 'guides/deep-linking' },
            { label: 'Security & Permissions', slug: 'guides/security-and-permissions' },
            { label: 'Single-File Apps', slug: 'guides/single-file-apps' },
            { label: 'Error Handling', slug: 'guides/error-handling' },
            { label: 'Cross-Platform', slug: 'guides/cross-platform' },
            { label: 'Migrate from Electron', slug: 'guides/migration-from-electron' },
            { label: 'Migrate from Neutralinojs', slug: 'guides/migration-from-neutralinojs' },
          ],
        },
        {
          label: 'Packaging',
          items: [
            { label: 'Overview', slug: 'guides/packaging' },
            { label: 'App Bundle (.app)', slug: 'guides/packaging/app-bundle' },
            { label: 'DMG Installer', slug: 'guides/packaging/dmg' },
            { label: 'AppImage', slug: 'guides/packaging/appimage' },
            { label: '.deb Package', slug: 'guides/packaging/deb' },
            { label: '.rpm Package', slug: 'guides/packaging/rpm' },
            { label: 'Code Signing', slug: 'guides/packaging/code-signing' },
            { label: 'App Icons', slug: 'guides/packaging/icons' },
            { label: 'App Identifier', slug: 'guides/packaging/app-id' },
          ],
        },
        {
          label: 'Auto-Updates',
          items: [
            { label: 'Overview', slug: 'guides/auto-updates' },
            { label: 'Setup', slug: 'guides/auto-updates/setup' },
            { label: 'Hosting Releases', slug: 'guides/auto-updates/hosting-releases' },
            { label: 'Update Manifest', slug: 'guides/auto-updates/update-manifest' },
            { label: 'Update Flow', slug: 'guides/auto-updates/update-flow' },
            { label: 'Security', slug: 'guides/auto-updates/security' },
            { label: 'UI Patterns', slug: 'guides/auto-updates/ui-patterns' },
            { label: 'Release Server', slug: 'guides/auto-updates/release-server' },
            { label: 'Signing Keys', slug: 'guides/auto-updates/signing-keys' },
            { label: 'CI/CD', slug: 'guides/auto-updates/ci-cd' },
            { label: 'GitHub Releases', slug: 'guides/auto-updates/github-releases' },
          ],
        },
        {
          label: 'API Reference',
          items: [
            { label: 'Overview', slug: 'api' },
            { label: 'Window', slug: 'api/window' },
            { label: 'File System', slug: 'api/fs' },
            { label: 'Dialogs', slug: 'api/dialog' },
            { label: 'Clipboard', slug: 'api/clipboard' },
            { label: 'System', slug: 'api/system' },
            { label: 'App', slug: 'api/app' },
            { label: 'Shell', slug: 'api/shell' },
            { label: 'Notifications', slug: 'api/notify' },
            { label: 'Tray', slug: 'api/tray' },
            { label: 'Menu', slug: 'api/menu' },
            { label: 'Store', slug: 'api/store' },
            { label: 'HTTP', slug: 'api/http' },
            { label: 'Process', slug: 'api/process' },
            { label: 'Shortcuts', slug: 'api/shortcuts' },
            { label: 'Updater', slug: 'api/updater' },
            { label: 'Events', slug: 'api/events' },
            { label: 'Configuration', slug: 'api/config' },
            { label: 'CLI', slug: 'api/cli' },
            { label: 'Error Codes', slug: 'api/errors' },
          ],
        },
        {
          label: 'Concepts',
          items: [
            { label: 'Architecture', slug: 'concepts/architecture' },
            { label: 'IPC Protocol', slug: 'concepts/ipc-protocol' },
            { label: 'Security Model', slug: 'concepts/security-model' },
            { label: 'Cross-Platform Rendering', slug: 'concepts/cross-platform-rendering' },
            { label: 'AI-Native Design', slug: 'concepts/ai-native-design' },
          ],
        },
        {
          label: 'LLM Integration',
          items: [
            { label: 'Overview', slug: 'llm' },
            { label: 'llms.txt Spec', slug: 'llm/llms-txt' },
            { label: 'Prompting Guide', slug: 'llm/prompting-guide' },
            { label: 'Cursor Rules', slug: 'llm/cursor-rules' },
            { label: 'AGENTS.md', slug: 'llm/agents-md' },
            { label: 'Example Prompts', slug: 'llm/example-prompts' },
            { label: 'Common Mistakes', slug: 'llm/common-mistakes' },
          ],
        },
      ],
    }),
  ],
});
