<script lang="ts">
  import { site } from '#lib/data/site.js';
  import Nav from './Nav.svelte';
</script>

<header class="site-header">
  <a class="brand" href="/">{site.name}</a>
  <Nav />
</header>

<style>
  .site-header {
    position: sticky;
    top: 0;
    z-index: var(--z-header);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-6);
    min-height: var(--header-h);
    padding: 0.75rem var(--gutter);
    border-bottom: 1px solid var(--color-line);
    background: var(--color-bg);
    background: color-mix(in srgb, var(--color-bg) 82%, transparent);
    -webkit-backdrop-filter: blur(12px);
    backdrop-filter: blur(12px);
  }

  .brand {
    color: var(--color-text);
    font-family: var(--font-display);
    font-size: var(--text-lg);
    line-height: 1;
    text-decoration: none;
  }

  .brand:hover {
    color: var(--color-accent);
  }

  @supports not (backdrop-filter: blur(1px)) {
    .site-header {
      background: var(--color-bg);
    }
  }

  @media (max-width: 900px) {
    .site-header {
      background: var(--color-bg);
      -webkit-backdrop-filter: none;
      backdrop-filter: none;
    }
  }

  @media (max-width: 720px) {
    .site-header {
      gap: var(--space-4);
      padding-block: 0.6rem;
    }

    .brand {
      font-size: var(--text-base);
    }
  }
</style>