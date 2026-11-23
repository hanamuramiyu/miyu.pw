<script lang="ts">
  import { ui } from '#lib/content/ui.js';
</script>

<a class="skip-link" href="#main">{ui.skipLink}</a>

<style>
  .skip-link {
    position: fixed;
    top: 0.75rem;
    left: 0.75rem;
    z-index: var(--z-overlay);
    padding: 0.6rem 1rem;
    border-radius: var(--radius-sm);
    background: var(--color-accent);
    color: var(--color-bg);
    font-size: var(--text-sm);
    font-weight: 700;
    text-decoration: none;
    transform: translateY(-160%);
    transition: transform var(--dur-fast) var(--ease-out);
  }

  .skip-link:focus-visible {
    transform: translateY(0);
    outline-color: var(--color-text);
  }
</style>