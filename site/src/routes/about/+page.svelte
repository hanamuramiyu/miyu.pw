<script lang="ts">
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Seo from '#lib/components/seo/Seo.svelte';
  import Hours from '#lib/components/ui/Hours.svelte';
  import ContactList from '#lib/components/ui/ContactList.svelte';

  let { data }: PageProps = $props();
</script>

<Seo
  title={`${ui.about.title} - ${site.name}`}
  description={site.description}
  path="/about"
  noindex={site.noindex}
/>

<section class="page" aria-labelledby="about-title">
  <h1 id="about-title" class="title">{ui.about.title}</h1>

  <div class="intro-grid">
    <div class="main">
      <div class="copy">
        {#each data.about.aboutLines as line (line)}
          <p>{line}</p>
        {/each}
      </div>

      <section class="block block--do" aria-labelledby="about-do-title">
        <h2 id="about-do-title" class="block-title">{ui.about.doTitle}</h2>
        <ul class="do-list">
          {#each data.about.aboutDo as item (item.what)}
            <li>
              <span class="do-what">{item.what}</span>
              <span class="do-stack">{item.stack}</span>
            </li>
          {/each}
        </ul>
      </section>
    </div>

    <div class="side">
      <dl class="facts">
        {#each data.about.aboutFacts as fact (fact.label)}
          <div class="fact">
            <dt>{fact.label}</dt>
            <dd>{fact.value}</dd>
          </div>
        {/each}
      </dl>

      <section class="block block--tools" aria-labelledby="about-tools-title">
        <h2 id="about-tools-title" class="block-title">{ui.about.toolsTitle}</h2>
        <dl class="tool-list">
          {#each data.about.aboutTools as group (group.label)}
            <div class="tool-group">
              <dt>{group.label}</dt>
              <dd>{group.items}</dd>
            </div>
          {/each}
        </dl>
      </section>
    </div>
  </div>
</section>

<section class="contact" aria-labelledby="about-contact-title">
  <div class="contact-grid">
    <div class="contact-copy">
      <h2 id="about-contact-title" class="contact-title">{ui.contact.title}</h2>
      <p class="contact-intro">{data.about.contactIntro}</p>
      <p class="nometa">
        {ui.contact.nometaBefore}<a
          href="https://nometa.xyz"
          target="_blank"
          rel="noopener noreferrer">{ui.contact.nometaLink}</a>{ui.contact.nometaAfter}
      </p>
      <div class="hours">
        <Hours />
      </div>
    </div>

    <ContactList />
  </div>
</section>

<style>
  .page {
    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
  }

  .title {
    font-size: var(--text-3xl);
    font-weight: 500;
    line-height: var(--leading-tight);
    margin-bottom: clamp(1.5rem, 3vw, 2.5rem);
  }

  .intro-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.5fr) minmax(12rem, 0.7fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .main,
  .side {
    display: grid;
    align-content: start;
    row-gap: clamp(1.5rem, 3vw, 2rem);
  }

  .block {
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-line);
  }

  .block--do {
    padding-top: 0;
    border-top: 0;
  }

  .copy p {
    max-width: var(--measure);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .copy p:first-child {
    color: var(--color-text);
  }

  .copy p + p {
    margin-top: var(--space-4);
  }

  .facts {
    display: grid;
    align-content: start;
    align-self: start;
    font-size: var(--text-sm);
  }

  .fact {
    display: grid;
    grid-template-columns: 7rem minmax(0, 1fr);
    gap: var(--space-3);
    padding-block: 0.65rem;
    border-top: 1px solid var(--color-line);
  }

  .fact:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  .fact dt {
    color: var(--color-text-muted);
  }

  .fact dd {
    margin: 0;
    color: var(--color-text);
  }

  .block-title {
    margin-bottom: var(--space-3);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: 500;
  }

  .do-list {
    display: grid;
    font-size: var(--text-sm);
  }

  .do-list li {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding-block: 0.9rem;
    border-top: 1px solid var(--color-line);
  }

  .do-list li:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  .do-what {
    color: var(--color-text);
    font-size: var(--text-lg);
    font-weight: 500;
    line-height: 1.3;
  }

  .do-stack {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    line-height: 1.5;
  }

  .tool-list {
    display: grid;
    font-size: var(--text-sm);
  }

  .tool-group {
    display: grid;
    grid-template-columns: 7rem minmax(0, 1fr);
    gap: var(--space-3);
    padding-block: 0.65rem;
    border-top: 1px solid var(--color-line);
  }

  .tool-group:last-child {
    border-bottom: 1px solid var(--color-line);
  }

  .tool-group dt {
    color: var(--color-text-muted);
  }

  .tool-group dd {
    margin: 0;
    color: var(--color-text);
    line-height: 1.5;
  }

  .contact {
    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
    border-top: 1px solid var(--color-line);
  }

  .contact-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .contact-title {
    margin-bottom: var(--space-4);
    font-size: var(--text-2xl);
    font-weight: 500;
  }

  .contact-intro {
    color: var(--color-text);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .nometa {
    margin-top: var(--space-4);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  .nometa a {
    color: var(--color-text);
  }

  .nometa a:hover {
    color: var(--color-accent);
  }

  .hours {
    margin-top: var(--space-6);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-line);
  }

  @media (max-width: 1000px) {
    .intro-grid,
    .contact-grid {
      grid-template-columns: minmax(0, 1fr);
      gap: var(--space-8);
    }

    .main,
    .side {
      display: contents;
    }

    .copy {
      order: 1;
    }

    .facts {
      order: 2;
    }

    .block--do {
      order: 3;
    }

    .block--tools {
      order: 4;
    }
  }

  @media (max-width: 600px) {
    .fact,
    .tool-group {
      grid-template-columns: 6.5rem minmax(0, 1fr);
    }
  }
</style>
