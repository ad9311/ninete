import { paraglideVitePlugin } from '@inlang/paraglide-js';
import tailwindcss from '@tailwindcss/vite';
import { svelteTesting } from '@testing-library/svelte/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	return {
		plugins: [
			tailwindcss(),
			sveltekit(),
			paraglideVitePlugin({
				project: './project.inlang',
				outdir: './src/lib/paraglide'
			})
		],
		test: {
			env: loadEnv(mode, process.cwd(), ''),
			workspace: [
				{
					extends: true,
					plugins: [svelteTesting()],
					test: {
						name: 'client',
						environment: 'jsdom',
						clearMocks: true,
						include: ['src/**/*.svelte.{test,spec}.{js,ts}'],
						exclude: ['src/lib/server/**'],
						setupFiles: ['./vitest-setup-client.ts']
					}
				},
				{
					extends: true,
					test: {
						name: 'server',
						environment: 'node',
						include: ['src/**/*.{test,spec}.{js,ts}'],
						exclude: ['src/**/*.svelte.{test,spec}.{js,ts}'],
						setupFiles: ['./vitest-setup-server.ts']
					}
				}
			]
		}
	};
});
