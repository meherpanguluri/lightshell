// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	integrations: [
		starlight({
			title: 'LightShell Docs',
			social: [
				{
					icon: 'github',
					label: 'GitHub',
					href: 'https://github.com/meherpanguluri/lightshell',
				},
			],
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
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'File System', slug: 'guides/file-system' },
						{ label: 'Menus & Tray', slug: 'guides/menus-tray' },
						{ label: 'Cross-Platform', slug: 'guides/cross-platform' },
					],
				},
				{
					label: 'API Reference',
					items: [
						{ label: 'Window', slug: 'api/window' },
						{ label: 'File System', slug: 'api/fs' },
						{ label: 'Dialogs', slug: 'api/dialog' },
						{ label: 'Clipboard', slug: 'api/clipboard' },
						{ label: 'System & App', slug: 'api/system' },
					],
				},
				{
					label: 'Concepts',
					items: [
						{ label: 'Architecture', slug: 'concepts/architecture' },
						{ label: 'IPC Protocol', slug: 'concepts/ipc' },
					],
				},
			],
		}),
	],
});
