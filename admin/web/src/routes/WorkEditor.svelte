<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { ApiError, api, errorNotice, type Notice } from '../lib/api';
  import Banner from '../lib/components/Banner.svelte';
  import PublishToSite, {
    type PublishPhase,
    type PublishProgress
  } from '../lib/components/PublishToSite.svelte';
  import { renderMarkdown } from '../lib/markdown';
  import { navigate } from '../lib/router';
  import type { Color, Font, Image, Link, Work, WorkInput } from '../lib/types';

  let {
    id,
    onNavigate,
    onUnauthorized
  }: {
    id: string | null;
    onNavigate: (to: string) => void;
    onUnauthorized: () => void;
  } = $props();

  const slugPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/;
  const hexPattern = /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/;
  const SITE_URL = (import.meta.env.VITE_SITE_URL as string | undefined) || 'https://miyu.pw';

  let title = $state('');
  let slug = $state('');
  let summary = $state('');
  let kind = $state('');
  let audience = $state('');
  let role = $state('');
  let client = $state('');
  let date = $state('');
  let updated = $state('');
  let featured = $state(false);
  let sortOrder = $state(0);
  let body = $state('');
  let published = $state(false);
  let tools = $state<string[]>([]);
  let links = $state<Link[]>([]);
  let colors = $state<Color[]>([]);
  let fonts = $state<Font[]>([]);
  let images = $state<Image[]>([]);

  let loading = $state(true);
  let notFound = $state(false);
  let busy = $state('');
  let error = $state<Notice | null>(null);
  let fieldErrors = $state<Record<string, string>>({});
  let slugTouched = $state(false);
  let textarea: HTMLTextAreaElement | undefined = $state();

  let saveHint = $state('');
  let siteUpdatedAt = $state('');
  let siteRemoval = $state(false);
  let publishPhase = $state<PublishPhase>('idle');
  let publishRemoval = $state(false);
  let publisher = $state<ReturnType<typeof PublishToSite>>();

  const isNew = $derived(id === null);
  const previewHtml = $derived(renderMarkdown(body));
  const publishBusy = $derived(publishPhase === 'starting' || publishPhase === 'running');
  const removingFromSite = $derived(publishBusy && publishRemoval);
  const disabled = $derived(busy !== '' || publishBusy);

  $effect(() => {
    if (isNew && !slugTouched) {
      slug = slugify(title);
    }
  });

  onMount(() => {
    void refreshSiteStatus();
    if (id === null) {
      const today = todayISO();
      date = today;
      updated = today;
      loading = false;
      return;
    }
    void load(id);
  });

  async function load(workId: string): Promise<void> {
    loading = true;
    error = null;
    notFound = false;

    try {
      applyWork(await api.getWork(workId));
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) {
        onUnauthorized();
        return;
      }
      if (failure instanceof ApiError && failure.status === 404) {
        notFound = true;
        return;
      }
      error = errorNotice(failure);
    } finally {
      loading = false;
    }
  }

  function applyWork(work: Work): void {
    title = work.title;
    slug = work.slug;
    summary = work.summary;
    kind = work.kind;
    audience = work.audience;
    role = work.role;
    client = work.client ?? '';
    date = work.date;
    updated = work.updated;
    featured = work.featured;
    sortOrder = work.sortOrder;
    body = work.bodyMd;
    published = work.published;
    tools = [...work.tools];
    links = work.links.map((link) => ({ ...link }));
    colors = work.colors.map((color) => ({ ...color }));
    fonts = work.fonts.map((font) => ({ ...font }));
    images = work.images.map((image) => ({ ...image }));
    slugTouched = true;
  }

  function slugify(value: string): string {
    return value
      .toLowerCase()
      .normalize('NFKD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  }

  function todayISO(): string {
    const now = new Date();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    return `${now.getFullYear()}-${month}-${day}`;
  }

  function isRealDate(value: string): boolean {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;

    const [year, month, day] = value.split('-').map(Number);
    const parsed = new Date(Date.UTC(year, month - 1, day));

    return (
      parsed.getUTCFullYear() === year &&
      parsed.getUTCMonth() === month - 1 &&
      parsed.getUTCDate() === day
    );
  }

  function isHttpUrl(value: string): boolean {
    try {
      const parsed = new URL(value);
      return (parsed.protocol === 'http:' || parsed.protocol === 'https:') && parsed.host !== '';
    } catch {
      return false;
    }
  }

  function validate(): boolean {
    const errors: Record<string, string> = {};

    if (title.trim() === '') errors.title = 'Title is required.';
    if (summary.trim() === '') errors.summary = 'Summary is required.';
    if (slug.trim() === '') {
      errors.slug = 'Slug is required.';
    } else if (!slugPattern.test(slug.trim())) {
      errors.slug = 'Use lowercase letters, digits and single hyphens.';
    }
    if (!isRealDate(date)) errors.date = 'Use a real date in YYYY-MM-DD format.';
    if (!isRealDate(updated)) errors.updated = 'Use a real date in YYYY-MM-DD format.';

    links.forEach((link, index) => {
      if (link.label.trim() === '' && link.href.trim() === '') return;
      if (!isHttpUrl(link.href.trim())) {
        errors[`links[${index}].href`] = 'Use an absolute http(s) URL.';
      }
    });

    colors.forEach((color, index) => {
      if (color.name.trim() === '' && color.hex.trim() === '') return;
      if (!hexPattern.test(color.hex.trim())) {
        errors[`colors[${index}].hex`] = 'Use a 3- or 6-digit hex colour.';
      }
    });

    fieldErrors = errors;
    return Object.keys(errors).length === 0;
  }

  function buildInput(nextPublished: boolean): WorkInput {
    return {
      slug: slug.trim(),
      title: title.trim(),
      summary: summary.trim(),
      bodyMd: body,
      kind: kind.trim(),
      audience: audience.trim(),
      role: role.trim(),
      client: client.trim() === '' ? null : client.trim(),
      date,
      updated,
      published: nextPublished,
      featured,
      sortOrder: Number.isFinite(sortOrder) ? Math.trunc(sortOrder) : 0,
      tools: tools.map((tool) => tool.trim()).filter((tool) => tool !== ''),
      links: links
        .filter((link) => link.label.trim() !== '' || link.href.trim() !== '')
        .map((link) => ({ label: link.label.trim(), href: link.href.trim() })),
      colors: colors
        .filter((color) => color.name.trim() !== '' || color.hex.trim() !== '')
        .map((color) => ({ name: color.name.trim(), hex: color.hex.trim() })),
      fonts: fonts
        .filter((font) => font.name.trim() !== '' || font.role.trim() !== '')
        .map((font) => ({ name: font.name.trim(), role: font.role.trim() })),
      images: images.map((image) => ({
        path: image.path,
        alt: image.alt.trim(),
        caption: image.caption.trim(),
        width: image.width,
        height: image.height,
        isCover: image.isCover
      }))
    };
  }

  async function saveChanges(): Promise<void> {
    if (!validate()) return;

    busy = 'save';
    error = null;
    saveHint = '';

    try {
      if (id === null) {
        const created = await api.createWork(buildInput(false));
        navigate(`/works/${created.id}`, true);
        return;
      }

      const saved = await api.updateWork(id, buildInput(published));
      published = saved.published;
      if (published) saveHint = 'Saved. Publish to site to update the site.';
    } catch (failure) {
      handleApiError(failure);
    } finally {
      busy = '';
    }
  }

  async function publishWork(): Promise<void> {
    if (!validate()) return;

    busy = 'publish';
    error = null;
    saveHint = '';

    try {
      if (id === null) {
        const created = await api.createWork(buildInput(true));
        navigate(`/works/${created.id}`, true);
        return;
      }

      await api.updateWork(id, buildInput(published));
      await api.publish(id, true);
      published = true;
      saveHint = 'Saved. Publish to site to update the site.';
    } catch (failure) {
      handleApiError(failure);
    } finally {
      busy = '';
    }
  }

  async function unpublishWork(): Promise<void> {
    if (id === null) return;
    if (!validate()) return;

    busy = 'unpublish';
    error = null;
    saveHint = '';

    try {
      await api.updateWork(id, buildInput(false));
      published = false;
      void publisher?.start({ removal: true });
    } catch (failure) {
      handleApiError(failure);
    } finally {
      busy = '';
    }
  }

  async function refreshSiteStatus(): Promise<void> {
    try {
      const status = await api.publishStatus();
      if (status.state === 'ok') {
        siteUpdatedAt = formatClock(status.finishedAt) || formatClock(status.startedAt);
        siteRemoval = false;
      }
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) {
        onUnauthorized();
      }
    }
  }

  function handlePublishProgress(progress: PublishProgress): void {
    publishPhase = progress.phase;
    publishRemoval = progress.removal;

    if (progress.phase === 'starting' || progress.phase === 'running') {
      saveHint = '';
    }

    if (progress.phase === 'ok') {
      siteUpdatedAt = progress.finishedAt;
      siteRemoval = progress.removal;
      saveHint = '';
    }
  }

  function formatClock(value: string | undefined): string {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
  }

  async function remove(): Promise<void> {
    if (id === null) return;
    if (!window.confirm(`Delete "${title.trim() || 'this work'}"? This cannot be undone.`)) return;

    busy = 'delete';
    error = null;

    try {
      await api.deleteWork(id);
      onNavigate('/');
    } catch (failure) {
      handleApiError(failure);
    } finally {
      busy = '';
    }
  }

  function handleApiError(failure: unknown): void {
    if (failure instanceof ApiError) {
      if (failure.status === 401) {
        onUnauthorized();
        return;
      }
      if (failure.field) {
        fieldErrors = { ...fieldErrors, [failure.field]: describeFieldError(failure.field, failure.message) };
        return;
      }
    }
    error = errorNotice(failure);
  }

  function describeFieldError(field: string, message: string): string {
    const linkMatch = /^links\[(\d+)\]\.href$/.exec(field);
    if (linkMatch) return `Link ${Number(linkMatch[1]) + 1}: ${message}`;

    const colorMatch = /^colors\[(\d+)\]\.hex$/.exec(field);
    if (colorMatch) return `Colour ${Number(colorMatch[1]) + 1}: ${message}`;

    const labels: Record<string, string> = {
      slug: 'Slug',
      title: 'Title',
      summary: 'Summary',
      date: 'Date',
      updated: 'Updated'
    };
    const label = labels[field];
    if (!label) return message;

    const trimmed = message.replace(/\.$/, '');
    return `${label} ${trimmed}.`;
  }

  function clear(field: string): void {
    if (!(field in fieldErrors)) return;
    const next = { ...fieldErrors };
    delete next[field];
    fieldErrors = next;
  }

  function linkError(index: number): string {
    return fieldErrors[`links[${index}].href`] ?? '';
  }

  function colorError(index: number): string {
    return fieldErrors[`colors[${index}].hex`] ?? '';
  }

  function addTool(): void {
    tools = [...tools, ''];
  }

  function removeTool(index: number): void {
    tools = tools.filter((_, position) => position !== index);
  }

  function addLink(): void {
    links = [...links, { label: '', href: '' }];
  }

  function removeLink(index: number): void {
    links = links.filter((_, position) => position !== index);
  }

  function addColor(): void {
    colors = [...colors, { name: '', hex: '' }];
  }

  function removeColor(index: number): void {
    colors = colors.filter((_, position) => position !== index);
  }

  function addFont(): void {
    fonts = [...fonts, { name: '', role: '' }];
  }

  function removeFont(index: number): void {
    fonts = fonts.filter((_, position) => position !== index);
  }

  async function surround(prefix: string, suffix: string, placeholder: string): Promise<void> {
    const element = textarea;
    if (!element) return;

    const start = element.selectionStart;
    const end = element.selectionEnd;
    const selected = body.slice(start, end) || placeholder;

    body = body.slice(0, start) + prefix + selected + suffix + body.slice(end);

    await tick();
    element.focus();
    element.setSelectionRange(start + prefix.length, start + prefix.length + selected.length);
  }

  async function prefixLines(prefix: string): Promise<void> {
    const element = textarea;
    if (!element) return;

    const start = element.selectionStart;
    const end = element.selectionEnd;
    const lineStart = body.lastIndexOf('\n', start - 1) + 1;
    const nextBreak = body.indexOf('\n', end);
    const lineEnd = nextBreak === -1 ? body.length : nextBreak;
    const block = body.slice(lineStart, lineEnd);
    const next = block
      .split('\n')
      .map((line) => prefix + line)
      .join('\n');

    body = body.slice(0, lineStart) + next + body.slice(lineEnd);

    await tick();
    element.focus();
    element.setSelectionRange(lineStart + prefix.length, lineStart + next.length);
  }

  async function insertLink(): Promise<void> {
    const element = textarea;
    if (!element) return;

    const start = element.selectionStart;
    const end = element.selectionEnd;
    const selected = body.slice(start, end) || 'link text';
    const url = 'https://';
    const text = `[${selected}](${url})`;

    body = body.slice(0, start) + text + body.slice(end);

    await tick();
    element.focus();
    const urlStart = start + selected.length + 3;
    element.setSelectionRange(urlStart, urlStart + url.length);
  }

  async function insertImage(): Promise<void> {
    const element = textarea;
    if (!element) return;

    const start = element.selectionStart;
    const end = element.selectionEnd;
    const selected = body.slice(start, end) || 'alt text';
    const text = `![${selected}](path/to/image.webp)`;

    body = body.slice(0, start) + text + body.slice(end);

    await tick();
    element.focus();
    element.setSelectionRange(start + 2, start + 2 + selected.length);
  }

  function keepSelection(event: MouseEvent): void {
    event.preventDefault();
  }
</script>

{#if loading}
  <p class="page muted">Loading...</p>
{:else if notFound}
  <div class="page">
    <div class="topbar">
      <h1>Not found</h1>
    </div>
    <p class="muted">This work does not exist or was deleted.</p>
    <a href="/" onclick={(event) => { event.preventDefault(); onNavigate('/'); }}>Back to works</a>
  </div>
{:else}
  <div class="page">
    <div class="topbar">
      <h1>{isNew ? 'New work' : 'Edit work'}</h1>
      <div class="editor-actions">
        <div class="editor-buttons">
          <span class="status" class:published>{published ? 'Published' : 'Draft'}</span>
          <button class="btn btn-small" type="button" onclick={() => onNavigate('/')}>Back</button>
          <button class="btn" type="button" disabled={disabled} onclick={() => void saveChanges()}>
            {busy === 'save' ? 'Saving...' : published ? 'Save changes' : 'Save draft'}
          </button>
          {#if published}
            <button class="btn" type="button" disabled={disabled} onclick={() => void unpublishWork()}>
              {busy === 'unpublish' ? 'Unpublishing...' : 'Unpublish'}
            </button>
          {:else}
            <button class="btn btn-primary" type="button" disabled={disabled} onclick={() => void publishWork()}>
              {busy === 'publish' ? 'Publishing...' : 'Publish'}
            </button>
          {/if}
          <PublishToSite
            bind:this={publisher}
            showButton={published}
            onUnauthorized={onUnauthorized}
            onProgress={handlePublishProgress}
          />
          {#if !isNew}
            <button
              class="btn btn-danger delete"
              type="button"
              disabled={disabled}
              onclick={() => void remove()}
            >
              {busy === 'delete' ? 'Deleting...' : 'Delete'}
            </button>
          {/if}
        </div>
        {#if removingFromSite}
          <p class="site-line">Removing from site...</p>
        {:else if siteUpdatedAt !== '' && (siteRemoval || published)}
          <p class="site-line">
            {#if siteRemoval}
              Removed from site at {siteUpdatedAt}
            {:else}
              Site updated at {siteUpdatedAt} ·
              <a href={SITE_URL} target="_blank" rel="noopener noreferrer">open site</a>
            {/if}
          </p>
        {/if}
        {#if saveHint !== ''}
          <p class="site-line">{saveHint}</p>
        {/if}
      </div>
    </div>

    {#if error !== null}
      <Banner title={error.title} detail={error.detail} />
    {/if}

    <div class="editor">
      <form class="form" onsubmit={(event) => event.preventDefault()}>
        <div class="field">
          <label class="label" for="work-title">Title</label>
          <input
            id="work-title"
            class="input"
            bind:value={title}
            oninput={() => clear('title')}
            autocomplete="off"
          />
          {#if fieldErrors.title}<p class="field-error">{fieldErrors.title}</p>{/if}
        </div>

        <div class="field">
          <label class="label" for="work-slug">Slug</label>
          <input
            id="work-slug"
            class="input"
            bind:value={slug}
            oninput={() => {
              slugTouched = true;
              clear('slug');
            }}
            autocomplete="off"
            spellcheck="false"
          />
          {#if fieldErrors.slug}<p class="field-error">{fieldErrors.slug}</p>{/if}
        </div>

        <div class="field">
          <label class="label" for="work-summary">Summary</label>
          <textarea
            id="work-summary"
            class="textarea"
            rows="3"
            bind:value={summary}
            oninput={() => clear('summary')}
          ></textarea>
          {#if fieldErrors.summary}<p class="field-error">{fieldErrors.summary}</p>{/if}
        </div>

        <div class="grid grid-2">
          <div class="field">
            <label class="label" for="work-kind">Kind</label>
            <input id="work-kind" class="input" bind:value={kind} autocomplete="off" />
          </div>
          <div class="field">
            <label class="label" for="work-audience">Audience</label>
            <input id="work-audience" class="input" bind:value={audience} autocomplete="off" />
          </div>
          <div class="field">
            <label class="label" for="work-role">Role</label>
            <input id="work-role" class="input" bind:value={role} autocomplete="off" />
          </div>
          <div class="field">
            <label class="label" for="work-client">Client</label>
            <input id="work-client" class="input" bind:value={client} autocomplete="off" />
          </div>
          <div class="field">
            <label class="label" for="work-date">Date</label>
            <input
              id="work-date"
              class="input"
              type="date"
              bind:value={date}
              oninput={() => clear('date')}
            />
            {#if fieldErrors.date}<p class="field-error">{fieldErrors.date}</p>{/if}
          </div>
          <div class="field">
            <label class="label" for="work-updated">Updated</label>
            <input
              id="work-updated"
              class="input"
              type="date"
              bind:value={updated}
              oninput={() => clear('updated')}
            />
            {#if fieldErrors.updated}<p class="field-error">{fieldErrors.updated}</p>{/if}
          </div>
          <div class="field">
            <label class="label" for="work-sort">Sort order</label>
            <input id="work-sort" class="input" type="number" step="1" bind:value={sortOrder} />
          </div>
          <div class="field">
            <span class="label">Flags</span>
            <label class="checkbox">
              <input type="checkbox" bind:checked={featured} />
              Featured
            </label>
          </div>
        </div>

        <section class="group">
          <div class="group-head">
            <h2>Tools</h2>
            <button class="btn btn-small" type="button" onclick={addTool}>Add tool</button>
          </div>
          {#if tools.length === 0}
            <p class="muted small">No tools.</p>
          {/if}
          {#each tools as _tool, index}
            <div class="repeat-row repeat-tool">
              <input
                class="input"
                bind:value={tools[index]}
                aria-label={`Tool ${index + 1}`}
                autocomplete="off"
              />
              <button class="btn btn-small" type="button" onclick={() => removeTool(index)}>
                Remove
              </button>
            </div>
          {/each}
        </section>

        <section class="group">
          <div class="group-head">
            <h2>Links</h2>
            <button class="btn btn-small" type="button" onclick={addLink}>Add link</button>
          </div>
          {#if links.length === 0}
            <p class="muted small">No links.</p>
          {/if}
          {#each links as _link, index}
            <div class="repeat-block">
              <div class="repeat-row repeat-link">
                <input
                  class="input"
                  placeholder="Label"
                  bind:value={links[index].label}
                  aria-label={`Link ${index + 1} label`}
                  autocomplete="off"
                />
                <input
                  class="input"
                  placeholder="https://"
                  bind:value={links[index].href}
                  oninput={() => clear(`links[${index}].href`)}
                  aria-label={`Link ${index + 1} URL`}
                  autocomplete="off"
                  spellcheck="false"
                />
                <button class="btn btn-small" type="button" onclick={() => removeLink(index)}>
                  Remove
                </button>
              </div>
              {#if linkError(index) !== ''}
                <p class="field-error">{linkError(index)}</p>
              {/if}
            </div>
          {/each}
        </section>

        <section class="group">
          <div class="group-head">
            <h2>Colours</h2>
            <button class="btn btn-small" type="button" onclick={addColor}>Add colour</button>
          </div>
          {#if colors.length === 0}
            <p class="muted small">No colours.</p>
          {/if}
          {#each colors as _color, index}
            <div class="repeat-block">
              <div class="repeat-row repeat-color">
                <input
                  class="input"
                  placeholder="Name"
                  bind:value={colors[index].name}
                  aria-label={`Colour ${index + 1} name`}
                  autocomplete="off"
                />
                <input
                  class="input"
                  placeholder="#d5c4b4"
                  bind:value={colors[index].hex}
                  oninput={() => clear(`colors[${index}].hex`)}
                  aria-label={`Colour ${index + 1} hex`}
                  autocomplete="off"
                  spellcheck="false"
                />
                <button class="btn btn-small" type="button" onclick={() => removeColor(index)}>
                  Remove
                </button>
              </div>
              {#if colorError(index) !== ''}
                <p class="field-error">{colorError(index)}</p>
              {/if}
            </div>
          {/each}
        </section>

        <section class="group">
          <div class="group-head">
            <h2>Fonts</h2>
            <button class="btn btn-small" type="button" onclick={addFont}>Add font</button>
          </div>
          {#if fonts.length === 0}
            <p class="muted small">No fonts.</p>
          {/if}
          {#each fonts as _font, index}
            <div class="repeat-row repeat-font">
              <input
                class="input"
                placeholder="Name"
                bind:value={fonts[index].name}
                aria-label={`Font ${index + 1} name`}
                autocomplete="off"
              />
              <input
                class="input"
                placeholder="Role"
                bind:value={fonts[index].role}
                aria-label={`Font ${index + 1} role`}
                autocomplete="off"
              />
              <button class="btn btn-small" type="button" onclick={() => removeFont(index)}>
                Remove
              </button>
            </div>
          {/each}
        </section>

        <section class="group">
          <div class="group-head">
            <h2>Body</h2>
            <span class="muted small">Markdown</span>
          </div>
          <div class="markdown">
            <div class="toolbar" role="toolbar" aria-label="Markdown formatting">
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void surround('**', '**', 'bold text')}
              >
                Bold
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void surround('*', '*', 'italic text')}
              >
                Italic
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void prefixLines('## ')}
              >
                H2
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void prefixLines('- ')}
              >
                List
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void prefixLines('> ')}
              >
                Quote
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void insertLink()}
              >
                Link
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void surround('`', '`', 'code')}
              >
                Code
              </button>
              <button
                class="btn btn-small"
                type="button"
                onmousedown={keepSelection}
                onclick={() => void insertImage()}
              >
                Image
              </button>
            </div>
            <textarea
              class="textarea body"
              rows="22"
              bind:this={textarea}
              bind:value={body}
              spellcheck="false"
              aria-label="Markdown body"
            ></textarea>
          </div>
        </section>
      </form>

      <aside class="preview">
        <div class="preview-head">
          <h2 class="preview-title">{title.trim() || 'Untitled'}</h2>
          {#if summary.trim() !== ''}
            <p class="preview-summary">{summary}</p>
          {/if}
        </div>
        {#if body.trim() === ''}
          <p class="muted">Nothing to preview yet.</p>
        {:else}
          <div class="prose">{@html previewHtml}</div>
        {/if}
      </aside>
    </div>
  </div>
{/if}

<style>
  .status {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    white-space: nowrap;
  }

  .status.published {
    color: var(--color-accent);
  }

  .editor-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: var(--space-1);
    min-width: 0;
  }

  .editor-buttons {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-2);
  }

  .editor-buttons .delete {
    margin-left: var(--space-2);
  }

  .site-line {
    margin: 0;
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    line-height: 1.3;
    text-align: right;
  }

  .site-line a {
    color: var(--color-accent);
  }

  .editor {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    align-items: start;
    gap: clamp(1.5rem, 3vw, 3rem);
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: var(--space-5);
    min-width: 0;
  }

  .grid {
    display: grid;
    gap: var(--space-4);
  }

  .grid-2 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .group {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-line);
  }

  .group-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .group-head h2 {
    margin: 0;
    font-size: var(--text-base);
    font-weight: 600;
  }

  .small {
    font-size: var(--text-xs);
  }

  .repeat-row {
    display: grid;
    align-items: center;
    gap: var(--space-2);
  }

  .repeat-block {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  .repeat-tool {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .repeat-link {
    grid-template-columns: minmax(0, 0.7fr) minmax(0, 1.3fr) auto;
  }

  .repeat-color {
    grid-template-columns: minmax(0, 1fr) 8rem auto;
  }

  .repeat-font {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
  }

  .markdown {
    border: 1px solid var(--color-line);
    border-radius: var(--radius);
  }

  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    padding: var(--space-1);
    border-bottom: 1px solid var(--color-line);
  }

  .toolbar .btn {
    border-color: transparent;
  }

  .toolbar .btn:hover {
    border-color: var(--color-line-strong);
  }

  .body {
    min-height: 26rem;
    border: 0;
    border-radius: 0;
    background: transparent;
    font-family: var(--font-mono);
    font-size: var(--text-xs);
    line-height: 1.6;
    resize: vertical;
  }

  .preview {
    position: sticky;
    top: var(--space-4);
    max-height: calc(100svh - 2 * var(--space-4));
    overflow-y: auto;
    padding-left: clamp(1rem, 2vw, 2rem);
    border-left: 1px solid var(--color-line);
    min-width: 0;
  }

  .preview-head {
    padding-bottom: var(--space-4);
    border-bottom: 1px solid var(--color-line);
  }

  .preview-title {
    margin: 0;
    font-size: var(--prose-2xl);
    font-weight: 600;
    line-height: 1.15;
  }

  .preview-summary {
    margin: var(--space-3) 0 0;
    color: var(--color-text-muted);
    font-size: var(--prose-base);
    line-height: 1.45;
  }

  .prose {
    max-width: var(--measure);
    padding-top: var(--space-4);
    font-size: var(--prose-base);
    line-height: 1.65;
  }

  .prose :global(h2) {
    margin-top: clamp(2.5rem, 5vw, 4rem);
    padding-top: var(--space-5);
    border-top: 1px solid var(--color-line);
    font-size: var(--prose-2xl);
    font-weight: 600;
    line-height: 1.15;
  }

  .prose :global(h2:first-child) {
    margin-top: 0;
    padding-top: 0;
    border-top: 0;
  }

  .prose :global(h3) {
    margin-top: var(--space-8);
    font-size: var(--prose-xl);
    font-weight: 600;
    line-height: 1.2;
  }

  .prose :global(h3:first-child) {
    margin-top: 0;
  }

  .prose :global(p) {
    margin-top: var(--space-4);
    color: var(--color-text-muted);
  }

  .prose :global(ul),
  .prose :global(ol) {
    margin-top: var(--space-4);
    padding-left: 1.25em;
    color: var(--color-text-muted);
  }

  .prose :global(ul) {
    list-style: disc;
  }

  .prose :global(ol) {
    list-style: decimal;
  }

  .prose :global(li) {
    margin-top: var(--space-2);
    padding-left: 0.25em;
  }

  .prose :global(li::marker) {
    color: var(--color-text-muted);
  }

  .prose :global(table) {
    width: 100%;
    margin-top: var(--space-5);
    margin-bottom: var(--space-2);
    border-collapse: collapse;
    border-top: 1px solid var(--color-line);
    border-bottom: 1px solid var(--color-line);
  }

  .prose :global(th) {
    padding: 0 0 var(--space-2);
    color: var(--color-text-muted);
    font-size: var(--text-sm);
    font-weight: 500;
    text-align: left;
  }

  .prose :global(td) {
    padding: var(--space-3) 0;
    border-top: 1px solid var(--color-line);
    vertical-align: top;
  }

  .prose :global(blockquote) {
    margin: clamp(2rem, 4vw, 3rem) 0;
    padding-left: var(--space-5);
    border-left: 2px solid var(--color-accent);
    color: var(--color-text);
    font-size: var(--prose-xl);
    font-style: normal;
    line-height: 1.65;
  }

  .prose :global(blockquote p) {
    margin-top: 0;
    color: inherit;
  }

  .prose :global(blockquote p + p) {
    margin-top: var(--space-3);
  }

  .prose :global(code) {
    padding: 0.1em 0.35em;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--color-line) 55%, transparent);
    color: var(--color-text);
    font-family: var(--font-mono);
    font-size: 0.9em;
  }

  .prose :global(pre) {
    margin-top: var(--space-4);
    padding: var(--space-3);
    overflow-x: auto;
    border: 1px solid var(--color-line);
    border-radius: var(--radius);
    background: var(--color-bg-raised);
  }

  .prose :global(pre code) {
    padding: 0;
    background: none;
  }

  .prose :global(a) {
    color: var(--color-accent);
  }

  .prose :global(img) {
    max-width: 100%;
    height: auto;
    margin-top: var(--space-4);
  }

  .prose :global(hr) {
    margin: var(--space-6) 0;
    border: 0;
    border-top: 1px solid var(--color-line);
  }

  @media (max-width: 960px) {
    .editor {
      grid-template-columns: minmax(0, 1fr);
    }

    .preview {
      position: static;
      max-height: none;
      padding-left: 0;
      padding-top: var(--space-5);
      border-left: 0;
      border-top: 1px solid var(--color-line);
    }
  }

  @media (max-width: 640px) {
    .grid-2,
    .repeat-link,
    .repeat-color,
    .repeat-font {
      grid-template-columns: minmax(0, 1fr);
    }

    .repeat-row .btn {
      justify-self: start;
    }
  }
</style>
