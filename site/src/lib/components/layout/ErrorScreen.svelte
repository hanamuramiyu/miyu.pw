<script lang="ts">
  import { site } from '#lib/data/site.js';
  import Seo from '#lib/components/seo/Seo.svelte';

  let {
    heading,
    message,
    homeLink,
    path = '/'
  }: {
    heading: string;
    message: string;
    homeLink: string;
    path?: string;
  } = $props();
</script>

<Seo title={`${heading} - ${site.name}`} description={message} {path} noindex={true} />

<section class="error-screen" aria-labelledby="error-title">
  <div class="content">
    <h1 id="error-title" class="heading">{heading}</h1>
    <p class="message">{message}</p>
    <a class="home" href="/">{homeLink}</a>
  </div>
</section>

<style>
  .error-screen {
    display: flex;
    align-items: center;
    min-height: min(36rem, calc(100svh - var(--header-h)));
    padding: var(--section-y) var(--gutter);
  }

  .content {
    max-width: var(--measure);
  }

  .heading {
    font-size: var(--text-3xl);
    line-height: var(--leading-tight);
  }

  .message {
    max-width: 46ch;
    margin-top: var(--space-5);
    color: var(--color-text-muted);
  }

  .home {
    display: inline-block;
    margin-top: var(--space-8);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .home:hover {
    color: var(--color-accent);
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  @media (max-width: 760px) {
    .error-screen {
      min-height: auto;
    }
  }
</style>