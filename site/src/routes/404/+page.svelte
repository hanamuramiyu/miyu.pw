<script lang="ts">
  import { ui } from '#lib/content/ui.js';
  import ErrorScreen from '#lib/components/layout/ErrorScreen.svelte';
</script>

<ErrorScreen
  heading={ui.notFound.heading}
  message={ui.notFound.message}
  homeLink={ui.notFound.homeLink}
  path="/404"
/>