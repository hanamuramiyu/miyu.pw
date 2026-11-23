<script lang="ts">
  import type { Snippet } from 'svelte';

  let { children }: { children: Snippet } = $props();
</script>

<div class="prose">
  {@render children()}
</div>

<style>
  .prose {
    max-width: var(--measure);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .prose :global(h2) {
    margin-top: clamp(2.5rem, 5vw, 4rem);
    padding-top: var(--space-5);
    border-top: 1px solid var(--color-line);
    font-family: var(--font-display);
    font-size: var(--text-2xl);
    font-weight: 500;
    line-height: var(--leading-heading);
  }

  .prose :global(h3) {
    margin-top: var(--space-8);
    font-family: var(--font-display);
    font-size: var(--text-xl);
    font-weight: 500;
    line-height: var(--leading-heading);
  }

  .prose :global(h2:first-child) {
    margin-top: 0;
    padding-top: 0;
    border-top: 0;
  }

  .prose :global(h3:first-child) {
    margin-top: 0;
  }

  .prose :global(p) {
    margin-top: var(--space-4);
    color: var(--color-text-muted);
  }

  .prose :global(ul),
  .prose :global(ol) {
    margin-top: var(--space-4);
    padding-left: 1.25em;
    color: var(--color-text-muted);
  }

  .prose :global(ul) {
    list-style: disc;
  }

  .prose :global(ol) {
    list-style: decimal;
  }

  .prose :global(li) {
    margin-top: var(--space-2);
    padding-left: 0.25em;
  }

  .prose :global(table) {
    width: 100%;
    margin-top: var(--space-5);
    margin-bottom: var(--space-2);
    border-collapse: collapse;
    border-top: 1px solid var(--color-line);
    border-bottom: 1px solid var(--color-line);
  }

  .prose :global(th) {
    padding: 0 0 var(--space-2);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: 500;
    text-align: left;
  }

  .prose :global(td) {
    padding: var(--space-3) 0;
    border-top: 1px solid var(--color-line);
    vertical-align: top;
  }

  .prose :global(blockquote) {
    margin: clamp(2rem, 4vw, 3rem) 0;
    padding-left: var(--space-5);
    border-left: 2px solid var(--color-accent);
    color: var(--color-text);
    font-size: var(--text-xl);
    font-style: normal;
    line-height: var(--leading-body);
  }

  .prose :global(blockquote p) {
    margin-top: 0;
    color: inherit;
  }

  .prose :global(blockquote p + p) {
    margin-top: var(--space-3);
  }

  .prose :global(li::marker) {
    color: var(--color-text-muted);
  }

  .prose :global(code) {
    padding: 0.1em 0.35em;
    border-radius: var(--radius-sm);
    background-color: color-mix(in srgb, var(--color-line) 55%, transparent);
    color: var(--color-text);
    font-size: 0.9em;
  }

  .prose :global(a) {
    color: var(--color-accent);
    text-decoration-color: color-mix(in srgb, var(--color-accent) 55%, transparent);
  }

  .prose :global(a:hover) {
    color: var(--color-accent-soft);
    text-decoration-color: currentColor;
  }
</style>
