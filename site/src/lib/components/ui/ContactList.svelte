<script lang="ts">
  import { onDestroy } from 'svelte';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Icon from '#lib/components/ui/Icon.svelte';

  type CopyStatus = 'idle' | 'copied' | 'failed';
  let copyStatus = $state<CopyStatus>('idle');
  let copyTimer: ReturnType<typeof setTimeout> | undefined;

  const copyLabel = $derived(
    copyStatus === 'copied'
      ? ui.contact.copied
      : copyStatus === 'failed'
        ? ui.contact.copyFailed
        : ui.contact.copy
  );

  async function copyDiscordHandle() {
    try {
      if (!navigator.clipboard) {
        throw new Error('Clipboard API unavailable');
      }
      await navigator.clipboard.writeText(site.contacts.discord.handle);
      copyStatus = 'copied';
    } catch {
      copyStatus = 'failed';
    }

    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copyStatus = 'idle';
    }, 2000);
  }

  onDestroy(() => clearTimeout(copyTimer));
</script>

<ul class="contact-list">
  <li>
    <a
      class="contact-row"
      href={site.contacts.telegram.href}
      target="_blank"
      rel="noopener noreferrer"
    >
      <span class="contact-label">{site.contacts.telegram.label}</span>
      <span class="contact-value">{site.contacts.telegram.handle}</span>
      <span class="contact-tail" aria-hidden="true">
        <Icon name="arrow-up-right" size={15} />
      </span>
    </a>
  </li>

  <li>
    <div class="contact-row">
      <span class="contact-label">{site.contacts.discord.label}</span>
      <button
        type="button"
        class="contact-value discord-value"
        onclick={copyDiscordHandle}
        aria-label={ui.contact.copyAria}>{site.contacts.discord.handle}</button
      >
      <span class="contact-tail">
        <span class="copy-hint" aria-live="polite">{copyLabel}</span>
      </span>
    </div>
  </li>

  <li>
    <a
      class="contact-row"
      href={site.contacts.github.href}
      target="_blank"
      rel="noopener noreferrer"
    >
      <span class="contact-label">{site.contacts.github.label}</span>
      <span class="contact-value">{site.contacts.github.handle}</span>
      <span class="contact-tail" aria-hidden="true">
        <Icon name="arrow-up-right" size={15} />
      </span>
    </a>
  </li>
</ul>

<style>
  .contact-list {
    display: flex;
    flex-direction: column;
  }

  .contact-row {
    display: grid;
    grid-template-columns: 7rem minmax(0, 1fr) 1.25rem;
    column-gap: var(--space-4);
    align-items: center;
    padding-block: var(--space-5);
    border-top: 1px solid var(--color-line);
    color: var(--color-text);
    text-decoration: none;
  }

  .contact-list li:last-child .contact-row {
    border-bottom: 1px solid var(--color-line);
  }

  .contact-label {
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    transition: color var(--dur-fast) var(--ease-out);
  }

  .contact-value {
    min-width: 0;
    font-size: var(--text-lg);
  }

  .discord-value {
    justify-self: start;
    padding: 0;
    border: 0;
    background: none;
    color: var(--color-text);
    font-family: inherit;
    font-size: var(--text-lg);
    line-height: inherit;
    text-align: left;
    user-select: all;
    cursor: pointer;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .discord-value:hover {
    color: var(--color-accent);
  }

  .contact-tail {
    justify-self: end;
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    color: var(--color-text-muted);
    transition: color var(--dur-fast) var(--ease-out);
  }

  .copy-hint {
    font-size: var(--text-xs);
    white-space: nowrap;
  }

  a.contact-row:hover .contact-label,
  a.contact-row:hover .contact-value,
  a.contact-row:hover .contact-tail {
    color: var(--color-accent);
  }

  @media (max-width: 600px) {
    .contact-row {
      grid-template-columns: 5.5rem minmax(0, 1fr) 1.25rem;
      column-gap: var(--space-3);
    }
  }
</style>
