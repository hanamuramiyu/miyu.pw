<script lang="ts">
  import { site } from '#lib/data/site.js';

  let {
    title,
    description,
    path = '/',
    image = '/og/site.webp',
    type = 'website',
    noindex = false
  }: {
    title: string;
    description: string;
    path?: string;
    image?: string;
    type?: 'website' | 'article';
    noindex?: boolean;
  } = $props();

  const canonical = $derived(canonicalUrl(path));
  const imageUrl = $derived(new URL(image, site.url).href);

  function canonicalUrl(pathname: string): string {
    const url = new URL(pathname, site.url);
    url.pathname = `${url.pathname.replace(/\/+$/, '')}/`;
    return url.href;
  }
</script>

<svelte:head>
  <title>{title}</title>
  <meta name="description" content={description} />
  <link rel="canonical" href={canonical} />
  {#if noindex}
    <meta name="robots" content="noindex" />
  {/if}

  <meta property="og:title" content={title} />
  <meta property="og:description" content={description} />
  <meta property="og:url" content={canonical} />
  <meta property="og:type" content={type} />
  <meta property="og:site_name" content={site.name} />
  <meta property="og:locale" content="en_US" />
  <meta property="og:image" content={imageUrl} />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta property="og:image:type" content="image/webp" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={title} />
  <meta name="twitter:description" content={description} />
</svelte:head>