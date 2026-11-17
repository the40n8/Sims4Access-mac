import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		// `pnpm dev`: proxy the API to a bridge on this Mac
		// (game running, or `ts4access-bridge --demo`).
		host: true,
		proxy: { '/ts4access.v1.CompanionService': 'http://127.0.0.1:47800' }
	}
});
