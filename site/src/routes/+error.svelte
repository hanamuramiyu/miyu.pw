<script lang="ts">
  import { page } from '$app/state';
  import { ui } from '#lib/content/ui.js';
  import ErrorScreen from '#lib/components/layout/ErrorScreen.svelte';

  const status = $derived(page.status || 500);
  const notFound = $derived(status === 404);
  const heading = $derived(notFound ? ui.notFound.heading : ui.error.heading);
  const message = $derived(notFound ? ui.notFound.message : ui.error.message);
  const path = $derived(page.url.pathname);
</script>

<ErrorScreen
  {heading}
  {message}
  homeLink={ui.error.homeLink}
  {path}
/>