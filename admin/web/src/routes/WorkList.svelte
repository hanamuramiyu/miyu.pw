<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiError, api, errorNotice, type Notice } from '../lib/api';
  import Banner from '../lib/components/Banner.svelte';
  import type { Me, WorkSummary } from '../lib/types';

  let {
    me,
    onNavigate,
    onUnauthorized,
    onSignedOut
  }: {
    me: Me;
    onNavigate: (to: string) => void;
    onUnauthorized: () => void;
    onSignedOut: () => void;
  } = $props();

  let works = $state<WorkSummary[]>([]);
  let showDrafts = $state(true);
  let loading = $state(true);
  let error = $state<Notice | null>(null);
  let busyId = $state('');
  let signingOut = $state(false);
  let requestId = 0;

  onMount(() => {
    void refresh();
  });

  async function refresh(): Promise<void> {
    const current = ++requestId;
    loading = true;
    error = null;

    try {
      const response = await api.listWorks(showDrafts);
      if (current !== requestId) return;
      works = response.works;
    } catch (failure) {
      if (current !== requestId) return;
      if (failure instanceof ApiError && failure.status === 401) {
        onUnauthorized();
        return;
      }
      error = errorNotice(failure);
    } finally {
      if (current === requestId) loading = false;
    }
  }

  async function togglePublish(work: WorkSummary): Promise<void> {
    busyId = work.id;
    error = null;
    try {
      await api.publish(work.id, !work.published);
      await refresh();
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) {
        onUnauthorized();
        return;
      }
      error = errorNotice(failure);
    } finally {
      busyId = '';
    }
  }

  async function remove(work: WorkSummary): Promise<void> {
    if (!window.confirm(`Delete "${work.title}"? This cannot be undone.`)) return;

    busyId = work.id;
    error = null;
    try {
      await api.deleteWork(work.id);
      await refresh();
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) {
        onUnauthorized();
        return;
      }
      error = errorNotice(failure);
    } finally {
      busyId = '';
    }
  }

  async function signOut(): Promise<void> {
    signingOut = true;
    error = null;
    try {
      await api.logout();
      onSignedOut();
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) {
        onSignedOut();
        return;
      }
      error = errorNotice(failure);
    } finally {
      signingOut = false;
    }
  }

  function open(event: MouseEvent, to: string): void {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
      return;
    }
    event.preventDefault();
    onNavigate(to);
  }
</script>

<div class="page">
  <div class="topbar">
    <h1>Works</h1>
    <div class="topbar-actions">
      <a class="btn btn-primary" href="/works/new" onclick={(event) => open(event, '/works/new')}>
        New work
      </a>
    </div>
  </div>

  <div class="filters">
    <label class="checkbox">
      <input
        type="checkbox"
        checked={showDrafts}
        onchange={(event) => {
          showDrafts = event.currentTarget.checked;
          void refresh();
        }}
      />
      Show drafts
    </label>
    <span class="count muted">{works.length} total</span>
  </div>

  {#if error !== null}
    <Banner title={error.title} detail={error.detail} />
  {/if}

  {#if loading && works.length === 0}
    <p class="muted">Loading...</p>
  {:else if works.length === 0}
    <p class="muted">{showDrafts ? 'No works yet.' : 'No published works.'}</p>
  {:else}
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Title</th>
            <th>Slug</th>
            <th>Status</th>
            <th>Date</th>
            <th class="actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each works as work (work.id)}
            <tr>
              <td class="title">
                <a href={`/works/${work.id}`} onclick={(event) => open(event, `/works/${work.id}`)}>
                  {work.title}
                </a>
              </td>
              <td class="slug">{work.slug}</td>
              <td>
                <span class:published={work.published} class="status">
                  {work.published ? 'Published' : 'Draft'}
                </span>
              </td>
              <td class="date">{work.date}</td>
              <td class="actions">
                <a
                  class="action"
                  href={`/works/${work.id}`}
                  onclick={(event) => open(event, `/works/${work.id}`)}
                >
                  Edit
                </a>
                <button
                  class="action"
                  type="button"
                  disabled={busyId === work.id}
                  onclick={() => void togglePublish(work)}
                >
                  {work.published ? 'Unpublish' : 'Publish'}
                </button>
                <button
                  class="action action-danger"
                  type="button"
                  disabled={busyId === work.id}
                  onclick={() => void remove(work)}
                >
                  Delete
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  <footer class="session">
    <span class="muted">Signed in as <strong>{me.login}</strong></span>
    <button class="btn btn-small" type="button" disabled={signingOut} onclick={() => void signOut()}>
      {signingOut ? 'Signing out...' : 'Sign out'}
    </button>
  </footer>
</div>

<style>
  .filters {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
  }

  .count {
    font-size: var(--text-xs);
  }

  .table-wrap {
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--text-sm);
  }

  th {
    padding: var(--space-2) var(--space-3) var(--space-2) 0;
    border-bottom: 1px solid var(--color-line-strong);
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    font-weight: 500;
    text-align: left;
    white-space: nowrap;
  }

  td {
    padding: var(--space-2) var(--space-3) var(--space-2) 0;
    border-bottom: 1px solid var(--color-line);
    vertical-align: middle;
  }

  td:last-child,
  th:last-child {
    padding-right: 0;
  }

  .title {
    min-width: 12rem;
  }

  .title a {
    color: var(--color-text);
    text-decoration: none;
  }

  .title a:hover {
    color: var(--color-accent);
    text-decoration: underline;
  }

  .slug,
  .date {
    color: var(--color-text-muted);
    white-space: nowrap;
  }

  .status {
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    white-space: nowrap;
  }

  .status.published {
    color: var(--color-accent);
  }

  .actions {
    text-align: right;
    white-space: nowrap;
  }

  .action {
    padding: 0;
    border: 0;
    background: none;
    color: var(--color-text-muted);
    font-size: var(--text-xs);
    line-height: 1.4;
    text-decoration: underline;
    text-decoration-color: color-mix(in srgb, var(--color-text-muted) 45%, transparent);
    text-underline-offset: 0.2em;
  }

  .action + .action {
    margin-left: var(--space-3);
  }

  .action:hover {
    color: var(--color-accent);
    text-decoration-color: currentColor;
  }

  .action:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .action-danger:hover {
    color: var(--color-danger);
  }

  .session {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-top: var(--space-8);
    padding-top: var(--space-3);
    border-top: 1px solid var(--color-line);
    font-size: var(--text-sm);
  }

  .session strong {
    color: var(--color-text);
    font-weight: 500;
  }
</style>
