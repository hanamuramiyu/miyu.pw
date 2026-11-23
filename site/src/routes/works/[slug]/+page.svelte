<script lang="ts">
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import { formatPeriod } from '#lib/utils/period.js';
  import Seo from '#lib/components/seo/Seo.svelte';
  import Prose from '#lib/components/ui/Prose.svelte';
  import Icon from '#lib/components/ui/Icon.svelte';

  let { data }: PageProps = $props();

  const work = $derived(data.work);
  const period = $derived(formatPeriod(work.date, work.updated, work.published));
  const images = $derived(work.images);
  const coverIndex = $derived(images.findIndex((image) => image.cover === true));
  const cover = $derived(coverIndex >= 0 ? images[coverIndex] : null);
  const gallery = $derived(images.filter((_, index) => index !== coverIndex));
</script>

<Seo
  title={`${work.title} - ${site.name}`}
  description={work.summary}
  path={`/works/${work.slug}`}
  type="article"
  noindex={site.noindex}
/>

<article class="work">
  <a class="back" href="/works/">{ui.work.all}</a>

  <div class="layout">
    <header class="header">
      <h1 class="title">{work.title}</h1>
      <p class="summary">{work.summary}</p>
      {#if cover}
        <figure class="cover">
          <img
            src={cover.src}
            alt={cover.alt}
            width={cover.width}
            height={cover.height}
            loading="eager"
            fetchpriority="high"
          />
          {#if cover.caption}
            <figcaption>{cover.caption}</figcaption>
          {/if}
        </figure>
      {/if}
    </header>

    <aside class="facts" aria-label={ui.work.factsLabel}>
      <section class="sidebar-block">
        <h2 class="sidebar-title">{ui.work.whatLabel}</h2>
        <dl class="fact-list">
          {#if work.kind}
            <div class="fact">
              <dt>{ui.work.kindLabel}</dt>
              <dd>{work.kind}</dd>
            </div>
          {/if}
          {#if work.audience}
            <div class="fact">
              <dt>{ui.work.audienceLabel}</dt>
              <dd>{work.audience}</dd>
            </div>
          {/if}
          <div class="fact">
            <dt>{ui.work.toolsLabel}</dt>
            <dd>{work.tools.join(' · ')}</dd>
          </div>
          <div class="fact">
            <dt>{ui.work.periodLabel}</dt>
            <dd class="period">{period}</dd>
          </div>
          {#if work.role}
            <div class="fact">
              <dt>{ui.work.roleLabel}</dt>
              <dd>{work.role}</dd>
            </div>
          {/if}
          {#if work.client}
            <div class="fact">
              <dt>{ui.work.clientLabel}</dt>
              <dd>{work.client}</dd>
            </div>
          {/if}
        </dl>
      </section>

      {#if work.headings.length > 0}
        <nav class="sidebar-block contents" aria-labelledby="contents-title">
          <h2 id="contents-title" class="sidebar-title">{ui.work.contentsLabel}</h2>
          <ul>
            {#each work.headings as heading (heading.id)}
              <li>
                <a href={`#${heading.id}`}>{heading.label}</a>
              </li>
            {/each}
          </ul>
        </nav>
      {/if}

      {#if work.links.length > 0}
        <section class="sidebar-block links" aria-labelledby="links-title">
          <h2 id="links-title" class="sidebar-title">{ui.work.linksLabel}</h2>
          <ul>
            {#each work.links as link (link.href)}
              <li>
                <a href={link.href} target="_blank" rel="noopener noreferrer">
                  {link.label}
                  <span class="arrow" aria-hidden="true">
                    <Icon name="arrow-up-right" size={13} />
                  </span>
                </a>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
    </aside>

    <div class="body">
      <Prose>{@html work.bodyHtml}</Prose>

      {#if gallery.length > 0}
        <section class="gallery" aria-label={ui.work.galleryLabel}>
          {#each gallery as image, index (index)}
            <figure class="shot">
              <img
                src={image.src}
                alt={image.alt}
                width={image.width}
                height={image.height}
                loading="lazy"
              />
              {#if image.caption}
                <figcaption>{image.caption}</figcaption>
              {/if}
            </figure>
          {/each}
        </section>
      {/if}

      {#if work.colors && work.colors.length > 0}
        <section class="palette" aria-labelledby="palette-title">
          <h2 id="palette-title" class="block-title">{ui.work.palette}</h2>
          <ul class="swatches">
            {#each work.colors as color, index (index)}
              <li class="swatch">
                <span class="dot" style:background-color={color.hex} aria-hidden="true"></span>
                <span class="swatch-name">{color.name}</span>
                <span class="swatch-hex">{color.hex}</span>
              </li>
            {/each}
          </ul>
        </section>
      {/if}

      {#if work.fonts && work.fonts.length > 0}
        <section class="type" aria-labelledby="type-title">
          <h2 id="type-title" class="block-title">{ui.work.type}</h2>
          <ul class="fonts">
            {#each work.fonts as font, index (index)}
              <li>{font.name} - {font.role}</li>
            {/each}
          </ul>
        </section>
      {/if}
    </div>
  </div>
</article>

<style>
  .work {
    --measure: 62ch;

    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
  }

  .back {
    display: inline-flex;
    margin-bottom: clamp(1.5rem, 3vw, 2.5rem);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .back:hover {
    color: var(--color-accent);
  }

  .layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(14rem, 18rem);
    grid-template-rows: auto 1fr;
    column-gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .header {
    grid-column: 1;
    grid-row: 1;
    max-width: none;
  }

  .body {
    grid-column: 1;
    grid-row: 2;
    min-width: 0;
  }

  .body :global(.prose) {
    margin-top: clamp(2rem, 4vw, 3rem);
  }

  .title {
    max-width: var(--measure);
    margin-bottom: 0;
    font-size: var(--text-3xl);
    font-weight: 500;
    line-height: var(--leading-tight);
  }

  .summary {
    max-width: 54ch;
    margin-top: var(--space-4);
    color: var(--color-text);
    font-size: var(--text-xl);
    line-height: 1.45;
  }

  .cover {
    margin: clamp(1.5rem, 3vw, 2.5rem) 0 0;
  }

  .cover img {
    display: block;
    width: 100%;
    height: auto;
  }

  .cover figcaption {
    max-width: var(--measure);
    margin-top: var(--space-2);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  .facts {
    grid-column: 2;
    grid-row: 1 / span 2;
    align-self: start;
    position: sticky;
    top: calc(var(--header-h) + var(--space-6));
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  .sidebar-block + .sidebar-block {
    margin-top: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-line);
  }

  .sidebar-title {
    margin-bottom: var(--space-3);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-xs);
    font-weight: 500;
  }

  .fact-list {
    display: grid;
  }

  .fact {
    display: grid;
    gap: 0.15rem;
    padding-block: 0.65rem;
    border-top: 1px solid var(--color-line);
  }

  .fact:first-child {
    padding-top: 0;
    border-top: 0;
  }

  .fact:last-child {
    padding-bottom: 0;
  }

  .fact dt {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  .fact dd {
    margin: 0;
    color: var(--color-text);
  }

  .period {
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .contents ul {
    display: grid;
    gap: 0.4rem;
  }

  .contents a {
    display: block;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.35;
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .contents a:hover {
    color: var(--color-accent);
  }

  .links ul {
    display: grid;
    gap: 0.35rem;
  }

  .links a {
    display: inline-flex;
    align-items: center;
    gap: 0.35em;
    color: var(--color-text);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .links a:hover {
    color: var(--color-accent);
  }

  .arrow {
    display: inline-flex;
    color: var(--color-text-muted);
  }

  .palette,
  .type {
    max-width: var(--measure);
    margin-top: clamp(2.5rem, 5vw, 4rem);
  }

  .gallery {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: clamp(1.5rem, 3vw, 2.5rem);
    margin-top: clamp(2.5rem, 5vw, 4rem);
  }

  .shot img {
    display: block;
    width: 100%;
    height: auto;
  }

  .shot figcaption {
    margin-top: var(--space-2);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  @media (min-width: 720px) {
    .gallery {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  .block-title {
    margin-bottom: var(--space-3);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: 500;
  }

  .swatches {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-4) clamp(1.25rem, 3vw, 2.5rem);
  }

  .swatch {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
  }

  .dot {
    width: 0.95rem;
    height: 0.95rem;
    border: 1px solid var(--color-line-strong);
    border-radius: 50%;
  }

  .swatch-hex {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  .fonts {
    display: grid;
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .fonts li {
    padding-block: 0.5rem;
    border-top: 1px solid var(--color-line);
  }

  .fonts li:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  @media (max-width: 900px) {
    .layout {
      display: flex;
      flex-direction: column;
      gap: var(--space-6);
    }

    .header {
      order: 1;
      max-width: none;
    }

    .facts {
      order: 2;
      position: static;
      padding-block: var(--space-5);
      border-block: 1px solid var(--color-line);
    }

    .body {
      order: 3;
      max-width: none;
    }
  }
</style>