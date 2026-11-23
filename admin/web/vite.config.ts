import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig, type Plugin } from 'vite';

const distDir = '../internal/webui/dist';

function keepDistPlaceholder(): Plugin {
  return {
    name: 'miyu-keep-dist-placeholder',
    generateBundle() {
      this.emitFile({ type: 'asset', fileName: '.gitkeep', source: '' });
    }
  };
}

export default defineConfig({
  plugins: [svelte(), keepDistPlaceholder()],
  build: {
    outDir: distDir,
    emptyOutDir: true
  },
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:8080',
      '/auth': 'http://127.0.0.1:8080'
    }
  }
});
