<script lang="ts">
  import type { Snippet } from 'svelte';
  import manropeLatin from '@fontsource-variable/manrope/files/manrope-latin-wght-normal.woff2?url';
  import comfortaaLatin from '@fontsource-variable/comfortaa/files/comfortaa-latin-wght-normal.woff2?url';
  import '#lib/styles/app.css';
  import SkipLink from '#lib/components/layout/SkipLink.svelte';
  import Header from '#lib/components/layout/Header.svelte';
  import Footer from '#lib/components/layout/Footer.svelte';

  let { children }: { children: Snippet } = $props();
</script>

<svelte:head>
  <link
    rel="preload"
    href={manropeLatin}
    as="font"
    type="font/woff2"
    crossorigin="anonymous"
  />
  <link
    rel="preload"
    href={comfortaaLatin}
    as="font"
    type="font/woff2"
    crossorigin="anonymous"
  />
</svelte:head>

<SkipLink />
<Header />
<main id="main" tabindex="-1">
  {@render children()}
</main>
<Footer />