<script lang="ts">
  let { title, detail = '' }: { title: string; detail?: string } = $props();
</script>

<p class="banner" role="alert">
  <span class="banner-title">{title}</span>
  {#if detail !== ''}
    <span class="banner-detail">{detail}</span>
  {/if}
</p>

<style>
  .banner {
    margin-bottom: var(--space-4);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--color-danger);
    border-radius: var(--radius);
    color: var(--color-danger);
    font-size: var(--text-sm);
  }

  .banner-title {
    font-weight: 500;
  }

  .banner-detail {
    display: block;
    margin-top: var(--space-1);
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }
</style>
