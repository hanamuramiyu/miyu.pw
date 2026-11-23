<script lang="ts">
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Seo from '#lib/components/seo/Seo.svelte';
  import WorkList from '#lib/components/works/WorkList.svelte';

  let { data }: PageProps = $props();
</script>

<Seo
  title={`${ui.work.title} - ${site.name}`}
  description={data.worksPage.intro}
  path="/works"
  noindex={site.noindex}
/>

<section class="page" aria-labelledby="works-title">
  <h1 id="works-title" class="title">{ui.work.title}</h1>
  <p class="intro">{data.worksPage.intro}</p>
  <WorkList works={data.works} headingLevel="h2" />
</section>

<style>
  .page {
    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
  }

  .title {
    font-size: var(--text-3xl);
    font-weight: 500;
    line-height: var(--leading-tight);
  }

  .intro {
    max-width: 54ch;
    margin-top: var(--space-4);
    margin-bottom: clamp(1.5rem, 3vw, 2.5rem);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }
</style>
