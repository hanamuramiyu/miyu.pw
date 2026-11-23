import { sveltekit } from '@sveltejs/kit/vite';
import adapter from '@sveltejs/adapter-static';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [
    sveltekit({
      adapter: adapter({
        pages: 'build',
        assets: 'build',
        fallback: undefined,
        precompress: true,
        strict: true
      }),
      paths: { relative: false },
      version: { pollInterval: 0 },
      prerender: {
        handleUnseenRoutes: ({ routes }) => {
          const unexpected = routes.filter((route) => route !== '/works/[slug]');
          if (unexpected.length > 0) {
            throw new Error(`Unseen prerender routes: ${unexpected.join(', ')}`);
          }
        }
      }
    })
  ]
});