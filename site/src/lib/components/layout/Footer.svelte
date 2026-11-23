<script lang="ts">
  import { site } from '#lib/data/site.js';
</script>

<footer class="site-footer">
  <p>© {site.year} {site.name}</p>
</footer>

<style>
  .site-footer {
    padding: var(--space-6) var(--gutter);
    border-top: 1px solid var(--color-line);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }
</style>