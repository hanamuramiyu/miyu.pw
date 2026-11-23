<script lang="ts">
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Seo from '#lib/components/seo/Seo.svelte';

  let { data }: PageProps = $props();
</script>

<Seo
  title={`${ui.process.title} - ${site.name}`}
  description={data.process.intro}
  path="/process"
  noindex={site.noindex}
/>

<section class="page" aria-labelledby="process-title">
  <h1 id="process-title" class="title">{ui.process.title}</h1>
  <p class="intro">{data.process.intro}</p>
</section>

<section class="section" aria-labelledby="process-stages-title">
  <h2 id="process-stages-title" class="section-title">{ui.process.stagesTitle}</h2>

  <ol class="stages">
    {#each data.process.stages as stage (stage.number)}
      <li class="stage">
        <div class="stage-head">
          <span class="stage-number">{stage.number}</span>
          <h3 class="stage-name">{stage.name}</h3>
        </div>
        <p class="stage-description">{stage.description}</p>
      </li>
    {/each}
  </ol>
</section>

<section class="section" aria-labelledby="process-design-title">
  <div class="section-grid">
    <h2 id="process-design-title" class="section-title">{ui.process.designVsMarkupTitle}</h2>
    <div class="section-content">
      {#each data.process.designVsMarkup as paragraph (paragraph)}
        <p>{paragraph}</p>
      {/each}
    </div>
  </div>
</section>

<section class="section" aria-labelledby="process-revisions-title">
  <div class="section-grid">
    <h2 id="process-revisions-title" class="section-title">{ui.process.revisionsTitle}</h2>
    <div class="section-content">
      <p>{data.process.revisions.intro}</p>

      <div class="list-grid">
        <div class="list-block">
          <h3 class="subheading">{ui.process.includedTitle}</h3>
          <ul class="dash-list">
            {#each data.process.revisions.included as item (item)}
              <li>{item}</li>
            {/each}
          </ul>
        </div>

        <div class="list-block">
          <h3 class="subheading">{ui.process.excludedTitle}</h3>
          <ul class="dash-list">
            {#each data.process.revisions.excluded as item (item)}
              <li>{item}</li>
            {/each}
          </ul>
        </div>
      </div>

      <p class="note">{data.process.revisions.note}</p>
    </div>
  </div>
</section>

<section class="section" aria-labelledby="process-rounds-title">
  <div class="section-grid">
    <h2 id="process-rounds-title" class="section-title">{ui.process.roundsTitle}</h2>
    <div class="section-content">
      {#each data.process.rounds as paragraph (paragraph)}
        <p>{paragraph}</p>
      {/each}
    </div>
  </div>
</section>

<section class="section" aria-labelledby="process-payments-title">
  <div class="section-grid">
    <h2 id="process-payments-title" class="section-title">{ui.process.paymentsTitle}</h2>
    <div class="section-content">
      {#each data.process.payments as paragraph (paragraph)}
        <p>{paragraph}</p>
      {/each}
    </div>
  </div>
</section>

<section class="section" aria-labelledby="process-cancellation-title">
  <div class="section-grid">
    <h2 id="process-cancellation-title" class="section-title">{ui.process.cancellationTitle}</h2>
    <div class="section-content">
      <ul class="dash-list">
        {#each data.process.cancellation as item (item)}
          <li>{item}</li>
        {/each}
      </ul>
    </div>
  </div>
</section>

<section class="section" aria-labelledby="process-guarantee-title">
  <div class="section-grid">
    <h2 id="process-guarantee-title" class="section-title">{ui.process.guaranteeTitle}</h2>
    <div class="section-content">
      <p>{data.process.guarantee}</p>
    </div>
  </div>
</section>

<p class="closing">{data.process.closing}</p>

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
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .section {
    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
    border-top: 1px solid var(--color-line);
  }

  .section-title {
    font-size: var(--text-2xl);
    font-weight: 500;
  }

  .stages {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: clamp(2rem, 5vw, 4rem);
    margin-top: var(--space-4);
  }

  .stage {
    padding-block: clamp(1.1rem, 2vw, 1.4rem);
    border-top: 1px solid var(--color-line);
  }

  .stage-head {
    display: flex;
    align-items: baseline;
    gap: var(--space-3);
  }

  .stage-number {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    line-height: 1;
    font-variant-numeric: tabular-nums;
  }

  .stage-name {
    font-size: var(--text-xl);
    font-weight: 500;
    line-height: 1.25;
  }

  .stage-description {
    margin-top: var(--space-2);
    font-size: var(--text-base);
    line-height: var(--leading-body);
  }

  .section-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .section-content {
    min-width: 0;
  }

  .section-content p {
    max-width: var(--measure);
    color: var(--color-text);
    font-size: var(--text-base);
    line-height: var(--leading-body);
  }

  .section-content p + p {
    margin-top: var(--space-4);
  }

  .list-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: clamp(1.5rem, 4vw, 3rem);
    margin-top: var(--space-6);
  }

  .subheading {
    margin-bottom: var(--space-2);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: 500;
  }

  .dash-list {
    display: grid;
    gap: var(--space-1);
    max-width: var(--measure);
    font-size: var(--text-base);
    line-height: 1.55;
  }

  .dash-list li {
    position: relative;
    padding-left: 1.4em;
  }

  .dash-list li::before {
    content: '-';
    position: absolute;
    left: 0;
    color: var(--color-text-muted);
  }

  .note {
    margin-top: var(--space-6);
  }

  .closing {
    padding: var(--space-6) var(--gutter) calc(var(--section-y) * 0.6);
    border-top: 1px solid var(--color-line);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  @media (max-width: 900px) {
    .stages,
    .section-grid,
    .list-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .section-grid {
      gap: var(--space-4);
    }
  }
</style>