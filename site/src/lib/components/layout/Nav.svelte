<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { tick } from 'svelte';
  import { site } from '#lib/data/site.js';
  import { ui } from '#lib/content/ui.js';
  import Icon from '#lib/components/ui/Icon.svelte';

  const uid = $props.id();
  const menuId = `nav-menu-${uid}`;
  let open = $state(false);
  let toggleButton: HTMLButtonElement | undefined;
  let firstLink: HTMLAnchorElement | undefined;

  function toggleMenu() {
    if (open) {
      closeMenu(true);
      return;
    }

    open = true;
    void tick().then(() => firstLink?.focus());
  }

  function closeMenu(returnFocus: boolean) {
    if (!open) return;
    open = false;
    if (returnFocus) toggleButton?.focus();
  }

  function handleLinkClick() {
    closeMenu(false);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && open) {
      event.preventDefault();
      closeMenu(true);
    }
  }

  $effect(() => {
    if (!open) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    return () => {
      document.body.style.overflow = previousOverflow;
    };
  });

  afterNavigate(() => {
    closeMenu(false);
  });
</script>

<svelte:window onkeydown={handleKeydown} />

<nav class="nav" aria-label={ui.nav.label}>
  <button
    type="button"
    class="toggle"
    aria-expanded={open}
    aria-controls={menuId}
    aria-label={open ? ui.nav.close : ui.nav.menu}
    bind:this={toggleButton}
    onclick={toggleMenu}
  >
    <Icon name="menu" size={20} />
  </button>

  <div id={menuId} class="menu" class:is-open={open}>
    <button type="button" class="close" aria-label={ui.nav.close} onclick={() => closeMenu(true)}>
      <Icon name="close" size={22} />
    </button>

    <div class="menu-inner">
      <a href="/#work" bind:this={firstLink} onclick={handleLinkClick}>{ui.nav.work}</a>
      <a href="/#about" onclick={handleLinkClick}>{ui.nav.about}</a>
      <a href="/process/" onclick={handleLinkClick}>{ui.nav.process}</a>
      <a href="/design/" onclick={handleLinkClick}>{ui.nav.design}</a>
      <a href="/#contact" onclick={handleLinkClick}>{ui.nav.contact}</a>
    </div>

    <p class="menu-note">
      {site.hours.days}, {site.hours.open}–{site.hours.close} {site.hours.zoneLabel}
    </p>
  </div>
</nav>

<style>
  .nav {
    display: flex;
    align-items: center;
  }

  .menu {
    display: flex;
    align-items: center;
  }

  .menu-inner {
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    gap: clamp(0.9rem, 2.4vw, 1.75rem);
    font-size: var(--text-sm);
    font-weight: 500;
  }

  .menu-inner a {
    padding-block: 0.4rem;
    color: var(--color-text-muted);
    text-decoration: none;
    transition: color var(--dur-fast) var(--ease-out);
  }

  .menu-inner a:hover {
    color: var(--color-accent);
  }

  .menu-inner a:focus-visible {
    color: var(--color-text);
  }

  .toggle,
  .close,
  .menu-note {
    display: none;
  }

  .toggle,
  .close {
    align-items: center;
    justify-content: center;
    padding: 0.5rem;
    border: 1px solid var(--color-line);
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--color-text);
    line-height: 0;
  }

  .toggle:hover,
  .close:hover {
    border-color: var(--color-accent);
    color: var(--color-accent);
  }

  @media (max-width: 1000px) {
    .menu-inner {
      gap: clamp(0.55rem, 1.5vw, 0.95rem);
      font-size: var(--text-xs);
    }

    .menu-inner a {
      padding-block: 0.5rem;
    }
  }

  @media (max-width: 900px) {
    .toggle {
      display: inline-flex;
    }

    .menu {
      display: none;
    }

    .menu.is-open {
      position: fixed;
      inset: 0;
      z-index: var(--z-overlay);
      display: flex;
      flex-direction: column;
      justify-content: center;
      gap: var(--space-10);
      padding: calc(var(--header-h) + var(--space-6)) var(--gutter) var(--space-10);
      background-color: var(--color-bg);
    }

    .menu.is-open .close {
      position: absolute;
      top: 0.85rem;
      right: var(--gutter);
      display: inline-flex;
    }

    .menu.is-open .menu-inner {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-2);
      font-family: var(--font-display);
      font-size: clamp(2rem, 12vw, 4.5rem);
      font-weight: 500;
      line-height: 1.05;
    }

    .menu.is-open .menu-inner a {
      padding-block: 0.15em;
      color: var(--color-text);
    }

    .menu.is-open .menu-inner a:hover {
      color: var(--color-accent);
    }

    .menu.is-open .menu-note {
      display: block;
      color: var(--color-text-muted);
      font-size: var(--text-sm);
    }
  }
</style>