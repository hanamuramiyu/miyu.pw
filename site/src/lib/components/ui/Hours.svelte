<script lang="ts">
  import { onMount } from 'svelte';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';

  const formatter = new Intl.DateTimeFormat('en-GB', {
    timeZone: site.hours.timeZone,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  });

  let clock = $state('');
  let stamp = $state('');

  onMount(() => {
    const tick = () => {
      const now = new Date();
      clock = formatter.format(now);
      stamp = now.toISOString();
    };

    tick();
    const id = setInterval(tick, 30_000);
    return () => clearInterval(id);
  });
</script>

<div class="hours">
  <p class="schedule">
    {site.hours.days}, {site.hours.open}–{site.hours.close} {site.hours.zoneLabel}
  </p>
  <p class="local">
    <span>{ui.hours.local}</span>
    {#if clock}
      <time datetime={stamp} aria-live="off">{clock}</time>
    {:else}
      <span>{site.hours.zoneLabel}</span>
    {/if}
  </p>
</div>

<style>
  .hours {
    display: grid;
    gap: 0.35rem;
    font-size: var(--text-sm);
  }

  .schedule {
    color: var(--color-text);
    font-variant-numeric: tabular-nums;
  }

  .local {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.4rem;
    color: var(--color-text-muted);
  }

  time {
    color: var(--color-text);
    font-variant-numeric: tabular-nums;
  }
</style>