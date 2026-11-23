<script lang="ts">
  import type { WorkLink } from '#lib/content/schema.js';
  import Icon from '#lib/components/ui/Icon.svelte';

  let { links, arrow = false }: { links: WorkLink[]; arrow?: boolean } = $props();
</script>

{#if links.length > 0}
  <ul class="work-links">
    {#each links as link (link.href)}
      <li>
        <a href={link.href} target="_blank" rel="noopener noreferrer">
          {link.label}
          {#if arrow}
            <span class="arrow" aria-hidden="true"><Icon name="arrow-up-right" size={13} /></span>
          {/if}
        </a>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .work-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem 1.1rem;
    margin-top: var(--space-4);
    font-size: var(--text-sm);
  }

  .work-links a {
    display: inline-flex;
    align-items: center;
    gap: 0.3em;
    color: var(--color-text-muted);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .work-links a:hover {
    color: var(--color-accent);
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  .arrow {
    display: inline-flex;
    color: currentColor;
  }
</style>
