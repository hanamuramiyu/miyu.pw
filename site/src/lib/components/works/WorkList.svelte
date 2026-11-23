<script lang="ts">
  import type { WorkListItem } from '#lib/content/schema.js';
  import { ui } from '#lib/content/ui.js';
  import { formatPeriod } from '#lib/utils/period.js';
  import WorkLinks from './WorkLinks.svelte';

  let {
    works,
    headingLevel = 'h3'
  }: { works: WorkListItem[]; headingLevel?: 'h2' | 'h3' } = $props();
</script>

{#if works.length === 0}
  <p class="empty">{ui.work.empty}</p>
{:else}
  <ul class="work-list">
    {#each works as work (work.slug)}
      <li class="work-item">
        <div class="work-main">
          <svelte:element this={headingLevel} class="title">
            <a href={`/works/${work.slug}/`}>{work.title}</a>
          </svelte:element>
          <p class="summary">{work.summary}</p>
          <WorkLinks links={work.links} />
        </div>

        <div class="work-meta">
          <p class="tools">{work.tools.join(' · ')}</p>
          <p class="period">{formatPeriod(work.date, work.updated, work.published)}</p>
        </div>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .empty {
    padding-block: var(--space-8);
    color: var(--color-text-muted);
  }

  .work-list {
    display: flex;
    flex-direction: column;
  }

  .work-item {
    display: grid;
    grid-template-columns: minmax(0, 1.6fr) minmax(9rem, 0.8fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
    padding-block: clamp(1.75rem, 3.4vw, 2.75rem);
    border-top: 1px solid var(--color-line);
  }

  .work-item:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  .title {
    font-size: var(--text-2xl);
    font-weight: 500;
    line-height: var(--leading-tight);
  }

  .title a {
    color: inherit;
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .work-item:hover .title a,
  .work-item:focus-within .title a {
    color: var(--color-accent);
  }

  .summary {
    margin-top: var(--space-3);
    color: var(--color-text-muted);
  }

  .work-meta {
    display: grid;
    gap: 0.35rem;
    justify-items: end;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-align: right;
  }

  .tools {
    line-height: 1.5;
  }

  .period {
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  @media (max-width: 820px) {
    .work-item {
      grid-template-columns: minmax(0, 1fr);
      gap: var(--space-4);
    }

    .work-meta {
      justify-items: start;
      text-align: left;
    }
  }
</style>
