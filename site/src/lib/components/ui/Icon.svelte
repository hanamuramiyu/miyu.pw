<script lang="ts">
  type IconName = 'arrow-up-right' | 'menu' | 'close';

  let { name, size = 18 }: { name: IconName; size?: number } = $props();
</script>

<svg
  class="icon"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="1.5"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
  focusable="false"
>
  {#if name === 'arrow-up-right'}
    <path d="M6 18 18 6" />
    <path d="M9 6h9v9" />
  {:else if name === 'menu'}
    <path d="M4 7h16" />
    <path d="M4 12h16" />
    <path d="M4 17h16" />
  {:else if name === 'close'}
    <path d="m6 6 12 12" />
    <path d="m18 6-12 12" />
  {/if}
</svg>

<style>
  .icon {
    display: inline-block;
    flex: none;
    vertical-align: -0.15em;
  }
</style>