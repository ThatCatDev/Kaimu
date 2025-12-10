// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	server: {
		port: 4322,
	},
	integrations: [
		starlight({
			title: 'Pulse',
			description: 'A modern project management tool for software teams',
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/thatcatdev/pulse' },
			],
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'getting-started/introduction' },
						{ label: 'Quick Start', slug: 'getting-started/quick-start' },
						{ label: 'Installation', slug: 'getting-started/installation' },
					],
				},
				{
					label: 'Configuration',
					items: [
						{ label: 'Environment Variables', slug: 'configuration/environment-variables' },
						{ label: 'Authentication (OIDC)', slug: 'configuration/authentication' },
						{ label: 'Database', slug: 'configuration/database' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Setting up Google Auth', slug: 'guides/google-auth' },
						{ label: 'Setting up Okta', slug: 'guides/okta-auth' },
						{ label: 'Self-Hosting', slug: 'guides/self-hosting' },
					],
				},
				{
					label: 'API Reference',
					autogenerate: { directory: 'api' },
				},
				{
					label: 'Development',
					items: [
						{ label: 'Architecture', slug: 'development/architecture' },
						{ label: 'Contributing', slug: 'development/contributing' },
					],
				},
			],
			customCss: ['./src/styles/custom.css'],
		}),
	],
});
