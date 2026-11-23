<script lang="ts">
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Seo from '#lib/components/seo/Seo.svelte';
  import Icon from '#lib/components/ui/Icon.svelte';

  let { data }: PageProps = $props();
</script>

<Seo
  title={`${ui.design.title} - ${site.name}`}
  description={data.design.intro}
  path="/design"
  noindex={site.noindex}
/>

<section class="page" aria-labelledby="design-title">
  <h1 id="design-title" class="title">{ui.design.title}</h1>
  <p class="intro">{data.design.intro}</p>
</section>

<section class="section" aria-labelledby="design-colour-title">
  <h2 id="design-colour-title">{data.design.colour.title}</h2>
  <p class="section-note">{data.design.colour.note}</p>

  <ul class="swatches">
    {#each data.design.colour.tokens as token (token.token)}
      <li
        class:ink-light={token.ink === 'light'}
        class:ink-dark={token.ink === 'dark'}
        class="swatch"
        style={`background-color: var(--color-${token.token})`}
      >
        <span class="swatch-name">--color-{token.token}</span>
        <span class="swatch-hex">{token.hex}</span>
        <span class="swatch-role">{token.role}</span>
      </li>
    {/each}
  </ul>
</section>

<section class="section" aria-labelledby="design-type-title">
  <h2 id="design-type-title">{data.design.type.title}</h2>
  <p class="section-note">{data.design.type.note}</p>

  <table class="font-table">
    <thead>
      <tr>
        <th scope="col">{data.design.type.columns.font}</th>
        <th scope="col">{data.design.type.columns.usage}</th>
        <th scope="col">{data.design.type.columns.sample}</th>
      </tr>
    </thead>
    <tbody>
      {#each data.design.type.fonts as font (font.name)}
        <tr>
          <th scope="row">
            <a
              class="font-link"
              href={font.url}
              target="_blank"
              rel="noopener noreferrer"
            >
              {font.name}
              <span class="arrow" aria-hidden="true">
                <Icon name="arrow-up-right" size={12} />
              </span>
            </a>
          </th>
          <td class="font-usage">{font.usage}</td>
          <td>
            <span
              class="font-sample"
              class:font-sample--display={font.family === 'display'}
            >
              {data.design.type.pair[font.family]}
            </span>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>

  <ul class="type-scale">
    {#each data.design.type.steps as step (step.token)}
      <li class="type-row">
        <p
          class="type-sample"
          class:type-sample--display={step.font === 'display'}
          style={`font-size: var(--text-${step.token})`}
        >
          {step.sample}
        </p>
        <div class="type-meta">
          <span class="type-token">--text-{step.token}</span>
          <span class="type-size">{step.size}</span>
          <span class="type-purpose">{step.purpose}</span>
        </div>
      </li>
    {/each}
  </ul>
</section>

<section class="section" aria-labelledby="design-layout-title">
  <h2 id="design-layout-title">{data.design.layout.title}</h2>
  <p class="section-note">{data.design.layout.note}</p>

  <div class="row-demo">
    <div class="row-demo-main">
      <span class="row-label">{data.design.layout.workRow.left}</span>
      <p class="row-demo-title">{data.design.layout.workRow.title}</p>
      <p class="row-demo-description">{data.design.layout.workRow.description}</p>
    </div>
    <div class="row-demo-side">
      <span class="row-label">{data.design.layout.workRow.right}</span>
      <p class="row-demo-meta">{data.design.layout.workRow.tools}</p>
      <p class="row-demo-meta">{data.design.layout.workRow.period}</p>
    </div>
  </div>

  <h3 class="subheading">{data.design.layout.sectionsTitle}</h3>
  <p class="section-note">{data.design.layout.sectionsNote}</p>
  <ul class="surfaces">
    {#each data.design.layout.sections as surface (surface.name)}
      <li
        class="surface"
        class:surface--raised={surface.raised}
        class:surface--base={!surface.raised}
      >
        <span class="surface-name">{surface.name}</span>
        <span class="surface-token">--color-{surface.token}</span>
      </li>
    {/each}
  </ul>
</section>

<section class="section" aria-labelledby="design-character-title">
  <h2 id="design-character-title">{data.design.character.title}</h2>
  <p class="section-note">{data.design.character.note}</p>

  <div class="examples">
    {#each data.design.character.examples as example (example.src)}
      <figure class="example">
        <img
          src={example.src}
          width={example.width}
          height={example.height}
          alt={example.alt}
          loading="lazy"
          decoding="async"
        />
        <figcaption>{example.label}</figcaption>
      </figure>
    {/each}
  </div>
</section>

<style>
  .page {
    padding: var(--section-y) var(--gutter) clamp(2rem, 4vw, 3rem);
  }

  .title {
    font-size: var(--text-3xl);
    font-weight: 500;
    line-height: var(--leading-tight);
  }

  .intro {
    max-width: 54ch;
    margin-top: var(--space-4);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .section {
    padding: clamp(2rem, 4vw, 3rem) var(--gutter);
    border-top: 1px solid var(--color-line);
  }

  .section h2 {
    font-size: var(--text-2xl);
    font-weight: 500;
  }

  .section-note {
    max-width: var(--measure);
    margin-top: var(--space-3);
    margin-bottom: clamp(1.5rem, 3vw, 2.5rem);
    color: var(--color-text-muted);
    font-size: var(--text-base);
    line-height: var(--leading-body);
  }

  .swatches {
    display: grid;
    gap: var(--space-3);
  }

  .swatch {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-2) var(--space-6);
    min-height: clamp(5rem, 8vw, 7.5rem);
    padding: var(--space-4) var(--space-5);
  }

  .ink-light {
    color: var(--color-text);
  }

  .ink-dark {
    color: var(--color-bg);
  }

  .swatch-name {
    font-size: var(--text-sm);
    font-weight: 500;
  }

  .swatch-hex {
    font-family: ui-monospace, 'SFMono-Regular', 'Cascadia Mono', Menlo, monospace;
    font-size: var(--text-sm);
  }

  .swatch-role {
    flex-basis: 100%;
    order: 3;
    font-size: var(--text-xs);
  }

  @media (min-width: 720px) {
    .swatch {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
      align-items: center;
    }

    .swatch-hex {
      justify-self: center;
    }

    .swatch-role {
      order: 0;
      flex-basis: auto;
      justify-self: end;
      text-align: right;
    }
  }

  .font-table {
    width: 100%;
    border-collapse: collapse;
    table-layout: fixed;
    margin-bottom: clamp(2rem, 4vw, 3rem);
    text-align: left;
  }

  .font-table th,
  .font-table td {
    padding-block: clamp(1rem, 2.5vw, 1.6rem);
    border-top: 1px solid var(--color-line);
    vertical-align: middle;
    text-align: left;
  }

  .font-table thead th {
    width: auto;
    padding-block: 0 var(--space-3);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-xs);
    font-weight: 500;
    border-top: 0;
  }

  .font-table thead th:first-child {
    width: 11rem;
  }

  .font-table tbody th {
    width: 11rem;
    font-weight: 500;
  }

  .font-table tbody tr:last-child th,
  .font-table tbody tr:last-child td {
    border-bottom: 1px solid var(--color-line);
  }

  .font-link {
    display: inline-flex;
    align-items: center;
    gap: 0.3em;
    line-height: calc(var(--text-2xl) * 1.2);
    color: var(--color-text-muted);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .font-link:hover {
    color: var(--color-accent);
  }

  .arrow {
    display: inline-flex;
    color: currentColor;
  }

  .font-usage {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  .font-sample {
    display: flex;
    align-items: center;
    font-family: var(--font-body);
    font-size: var(--text-2xl);
    line-height: 1.2;
  }

  .font-sample--display {
    font-family: var(--font-display);
    font-weight: 500;
    line-height: 1.2;
  }

  .type-scale {
    display: grid;
  }

  .type-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(9rem, 14rem);
    gap: clamp(1rem, 4vw, 3rem);
    align-items: stretch;
    padding-block: clamp(1rem, 2.5vw, 1.75rem);
    border-top: 1px solid var(--color-line);
  }

  .type-row:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  .type-sample {
    display: flex;
    align-items: center;
    min-width: 0;
    overflow-wrap: break-word;
    font-family: var(--font-body);
    line-height: 1.2;
    text-wrap: balance;
  }

  .type-sample--display {
    font-family: var(--font-display);
    font-weight: 500;
    line-height: 1.1;
  }

  .type-meta {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 0.15rem;
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    line-height: 1.5;
  }

  .type-token {
    color: var(--color-text);
    font-family: ui-monospace, 'SFMono-Regular', 'Cascadia Mono', Menlo, monospace;
  }

  .row-demo {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(10rem, 18rem);
    gap: clamp(1.5rem, 4vw, 3rem);
    padding-block: clamp(1.5rem, 3vw, 2.25rem);
    border-top: 1px solid var(--color-line);
    border-bottom: 1px solid var(--color-line);
  }

  .row-demo-main,
  .row-demo-side {
    display: grid;
    align-content: start;
    gap: var(--space-2);
    min-width: 0;
  }

  .row-label {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    line-height: 1.4;
  }

  .row-demo-title {
    color: var(--color-text-muted);
    font-family: var(--font-display);
    font-size: var(--text-2xl);
    font-weight: 500;
    line-height: 1.2;
  }

  .row-demo-description,
  .row-demo-meta {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  .subheading {
    margin-top: clamp(2rem, 4vw, 3rem);
    margin-bottom: var(--space-3);
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: 500;
    color: var(--color-text-muted);
  }

  .surfaces {
    display: grid;
    margin-inline: calc(-1 * var(--gutter));
  }

  .surface {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    min-height: 4rem;
    padding: var(--space-3) var(--gutter);
    border-block: 1px solid var(--color-line);
  }

  .surface + .surface {
    margin-top: -1px;
  }

  .surface--base {
    background-color: var(--color-bg);
  }

  .surface--raised {
    background-color: var(--color-bg-raised);
  }

  .surface-name {
    color: var(--color-text-muted);
    font-family: var(--font-display);
    font-size: var(--text-base);
  }

  .surface-token {
    color: var(--color-text-muted);
    font-family: ui-monospace, 'SFMono-Regular', 'Cascadia Mono', Menlo, monospace;
    font-size: var(--text-xs);
    text-align: right;
  }

  .examples {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .example {
    display: grid;
    gap: var(--space-2);
  }

  .example img {
    height: auto;
    border: 1px solid var(--color-line);
  }

  .example:first-child img {
    width: 192px;
    height: 192px;
  }

  .example:last-child img {
    width: 100%;
    max-width: 640px;
  }

  .example figcaption {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
  }

  @media (max-width: 720px) {
    .type-row,
    .row-demo {
      grid-template-columns: minmax(0, 1fr);
    }

    .font-table {
      table-layout: auto;
    }

    .font-table thead {
      position: absolute;
      width: 1px;
      height: 1px;
      padding: 0;
      margin: -1px;
      overflow: hidden;
      clip-path: inset(50%);
      white-space: nowrap;
      border: 0;
    }

    .font-table tbody tr {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 0 var(--space-4);
      padding-block: var(--space-4);
      border-top: 1px solid var(--color-line);
    }

    .font-table tbody tr:last-child {
      border-bottom: 1px solid var(--color-line);
    }

    .font-table tbody th,
    .font-table tbody td {
      width: auto;
      padding: 0;
      border: 0;
    }

    .font-usage {
      grid-column: 1 / -1;
      grid-row: 2;
    }

    .font-table tbody td:last-child {
      grid-column: 2;
      grid-row: 1;
      justify-self: end;
    }

    .examples {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>