<script lang="ts">
  import { onMount } from 'svelte';
  import type { PageProps } from './$types';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Seo from '#lib/components/seo/Seo.svelte';
  import ContactList from '#lib/components/ui/ContactList.svelte';
  import WorkList from '#lib/components/works/WorkList.svelte';

  let { data }: PageProps = $props();

  const formatter = new Intl.DateTimeFormat('en-GB', {
    timeZone: site.hours.timeZone,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  });

  let clock = $state('');

  onMount(() => {
    const tick = () => {
      clock = formatter.format(new Date());
    };

    tick();
    const id = setInterval(tick, 30_000);
    return () => clearInterval(id);
  });
</script>

<Seo
  title={`${site.name} - ${site.role}`}
  description={site.description}
  path="/"
  noindex={site.noindex}
/>

<section class="hero" aria-labelledby="hero-title">
  <div class="identity">
    <h1 id="hero-title" class="wordmark">
      <span>hanamura</span>
      <span>miyu</span>
    </h1>

    <p class="role">{site.role}</p>
    <p class="hero-tagline">{data.home.heroTagline}</p>
    <p class="hero-intro">{data.home.heroIntro}</p>
  </div>

  <div class="hero-meta">
    <p class="meta-time">
      <span>{site.hours.days}, {site.hours.open}–{site.hours.close} {site.hours.zoneLabel}</span>
      <span class="meta-local">{ui.hours.local} <time>{clock || site.hours.zoneLabel}</time></span>
    </p>
  </div>
</section>

<section id="work" class="section section--raised" aria-labelledby="work-title">
  <h2 id="work-title" class="section-title">
    {ui.work.title}
    <a class="all-link" href="/works/">{ui.work.all}</a>
  </h2>
  <p class="section-intro">{data.home.workIntro}</p>
  <WorkList works={data.works} />
</section>

<section id="about" class="section" aria-labelledby="about-title">
  <h2 id="about-title" class="section-title">{ui.about.title}</h2>

  <div class="about-grid">
    <div class="about-main">
      <div class="about-copy">
        {#each data.home.aboutLines as line (line)}
          <p>{line}</p>
        {/each}
      </div>

      <section class="about-block about-block--do" aria-labelledby="about-do-title">
        <h3 id="about-do-title" class="block-title">{ui.about.doTitle}</h3>
        <ul class="do-list">
          {#each data.home.aboutDo as item (item.what)}
            <li>
              <span class="do-what">{item.what}</span>
              <span class="do-stack">{item.stack}</span>
            </li>
          {/each}
        </ul>
      </section>
    </div>

    <div class="about-side">
      <dl class="facts">
        {#each data.home.aboutFacts as fact (fact.label)}
          <div class="fact">
            <dt>{fact.label}</dt>
            <dd>{fact.value}</dd>
          </div>
        {/each}
      </dl>

      <section class="about-block about-block--tools" aria-labelledby="about-tools-title">
        <h3 id="about-tools-title" class="block-title">{ui.about.toolsTitle}</h3>
        <dl class="tool-list">
          {#each data.home.aboutTools as group (group.label)}
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

<section id="contact" class="section section--raised" aria-labelledby="contact-title">
  <h2 id="contact-title" class="section-title">{ui.contact.title}</h2>

  <div class="contact-grid">
    <div class="contact-copy">
      <p class="contact-intro">{data.home.contactIntro}</p>
      <p class="nometa">
        {ui.contact.nometaBefore}<a
          href="https://nometa.xyz"
          target="_blank"
          rel="noopener noreferrer">{ui.contact.nometaLink}</a>{ui.contact.nometaAfter}
      </p>
      <p class="contact-location">{data.home.contactLocation}</p>
    </div>

    <ContactList />
  </div>
</section>

<style>
  .hero {
    display: flex;
    flex-direction: column;
    justify-content: center;
    min-height: calc(100svh - var(--header-h));
    padding-block: clamp(2rem, 9vh, 5rem);
    padding-inline: var(--gutter);
  }

  .identity {
    min-width: 0;
    margin-block: auto;
  }

  .wordmark {
    display: flex;
    flex-direction: column;
    font-size: var(--text-display);
    font-weight: 500;
    line-height: 1;
  }

  .wordmark span {
    display: block;
  }

  .role {
    margin-top: var(--space-4);
    color: var(--color-text);
    font-size: var(--text-xl);
    line-height: var(--leading-body);
  }

  .hero-tagline {
    max-width: 48ch;
    margin-top: var(--space-1);
    color: var(--color-text);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .hero-intro {
    max-width: 52ch;
    margin-top: var(--space-3);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .hero-meta {
    display: flex;
    justify-content: flex-end;
    align-items: baseline;
    gap: var(--space-4);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-line);
    margin-top: auto;
  }

  .meta-time {
    font-size: var(--text-sm);
    color: var(--color-text-muted);
    line-height: 1.4;
  }

  .meta-time {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: var(--space-4);
  }

  .meta-local time {
    color: var(--color-text);
    font-variant-numeric: tabular-nums;
  }

  .section {
    padding: var(--section-y) var(--gutter) calc(var(--section-y) * 0.6);
    border-top: 1px solid var(--color-line);
  }

  .section--raised {
    background-color: var(--color-bg-raised);
  }

  .section-title {
    margin-bottom: clamp(1.25rem, 2.5vw, 2rem);
    font-size: var(--text-3xl);
    font-weight: 500;
    line-height: var(--leading-tight);
  }

  .all-link {
    margin-left: var(--space-4);
    color: var(--color-text-muted);
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: 500;
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .all-link:hover {
    color: var(--color-accent);
  }

  .section-intro {
    max-width: 54ch;
    margin-bottom: clamp(1.5rem, 3vw, 2.5rem);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .about-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.5fr) minmax(12rem, 0.7fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
  }

  .about-main,
  .about-side {
    display: grid;
    align-content: start;
    row-gap: clamp(1.5rem, 3vw, 2rem);
  }

  .about-block {
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-line);
  }

  .about-block--do {
    padding-top: 0;
    border-top: 0;
  }

  .about-copy p {
    max-width: var(--measure);
    color: var(--color-text-muted);
    font-size: var(--text-lg);
    line-height: var(--leading-body);
  }

  .about-copy p:first-child {
    color: var(--color-text);
  }

  .about-copy p + p {
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

  .contact-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
    gap: clamp(1.5rem, 4vw, 3rem);
    align-items: start;
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

  .contact-location {
    margin-top: var(--space-4);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
  }

  @media (max-width: 760px) {
    .hero-meta {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-2);
    }

    .meta-time {
      justify-content: flex-start;
    }
  }

  @media (max-width: 1000px) {
    .about-grid,
    .contact-grid {
      grid-template-columns: minmax(0, 1fr);
    }

    .about-grid,
    .contact-grid {
      gap: var(--space-8);
    }

    .about-main,
    .about-side {
      display: contents;
    }

    .about-copy {
      order: 1;
    }

    .facts {
      order: 2;
    }

    .about-block--do {
      order: 3;
    }

    .about-block--tools {
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
