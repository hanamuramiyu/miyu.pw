<script module lang="ts">
  export type PublishPhase = 'idle' | 'starting' | 'running' | 'ok' | 'failed';

  export interface PublishProgress {
    phase: PublishPhase;
    removal: boolean;
    finishedAt: string;
    error: string;
  }
</script>

<script lang="ts">
  import { onDestroy } from 'svelte';
  import { ApiError, api, errorNotice } from '../api';
  import Banner from './Banner.svelte';
  import type { PublishStatus } from '../types';

  let {
    onUnauthorized,
    onProgress,
    showButton = true
  }: {
    onUnauthorized: () => void;
    onProgress?: (progress: PublishProgress) => void;
    showButton?: boolean;
  } = $props();

  const POLL_MS = 2000;
  const TIMEOUT_MS = 120000;
  const FAILURE_HINT = 'The site was not changed. Fix the work and publish again.';
  const REMOVAL_WARNING = 'The work is a draft, but the site still shows the old version.';
  const MAX_FAILURE_LENGTH = 300;

  let phase = $state<PublishPhase>('idle');
  let failure = $state('');
  let removal = $state(false);
  let finishedAt = $state('');
  let timer: ReturnType<typeof setTimeout> | undefined;
  let destroyed = false;

  const busy = $derived(phase === 'starting' || phase === 'running');
  const failureDetail = $derived(
    `${failureText()}${removal ? ` ${REMOVAL_WARNING}` : ''}`
  );

  onDestroy(() => {
    destroyed = true;
    clearTimer();
  });

  function failureText(): string {
    const detail = sanitize(failure);
    return detail === '' ? FAILURE_HINT : `${detail} ${FAILURE_HINT}`;
  }

  function sanitize(text: string): string {
    return text
      .replace(/\x1B\[[0-9;]*[mK]/g, '')
      .replace(/\s+/g, ' ')
      .trim()
      .slice(0, MAX_FAILURE_LENGTH);
  }

  function notify(): void {
    onProgress?.({ phase, removal, finishedAt, error: failure });
  }

  function clearTimer(): void {
    if (timer !== undefined) {
      clearTimeout(timer);
      timer = undefined;
    }
  }

  function statusKey(status: PublishStatus): string {
    return [status.state, status.startedAt ?? '', status.finishedAt ?? '', status.release ?? ''].join('|');
  }

  function isFresh(status: PublishStatus, baseline: string): boolean {
    return statusKey(status) !== baseline;
  }

  function formatTime(value: string | undefined): string {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
  }

  export async function start(options: { removal?: boolean } = {}): Promise<void> {
    if (busy) return;

    clearTimer();
    phase = 'starting';
    failure = '';
    finishedAt = '';
    removal = options.removal === true;
    notify();

    let baseline = '';
    try {
      const previous = await api.publishStatus();
      if (destroyed) return;
      baseline = previous.state === 'idle' ? '' : statusKey(previous);
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        phase = 'idle';
        notify();
        return;
      }
    }

    try {
      await api.publishSite();
      if (destroyed) return;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        phase = 'idle';
        notify();
        return;
      }
      const notice = errorNotice(error);
      phase = 'failed';
      failure = sanitize(notice.detail === '' ? notice.title : `${notice.title}. ${notice.detail}`);
      notify();
      return;
    }

    phase = 'running';
    notify();
    void poll(baseline, Date.now() + TIMEOUT_MS, { sawUpdate: false });
  }

  async function poll(
    baseline: string,
    deadline: number,
    progress: { sawUpdate: boolean }
  ): Promise<void> {
    if (destroyed) return;

    try {
      const status = await api.publishStatus();
      if (destroyed) return;

      if (statusKey(status) !== baseline) {
        progress.sawUpdate = true;
      }

      if (status.state === 'ok' && isFresh(status, baseline)) {
        phase = 'ok';
        finishedAt = formatTime(status.finishedAt) || formatTime(status.startedAt) || formatTime(new Date().toISOString());
        notify();
        return;
      }

      if (status.state === 'failed' && isFresh(status, baseline)) {
        phase = 'failed';
        failure = sanitize(status.error?.trim() || 'Publish failed.');
        notify();
        return;
      }
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        onUnauthorized();
        phase = 'idle';
        notify();
        return;
      }
    }

    if (Date.now() >= deadline) {
      phase = 'failed';
      failure = progress.sawUpdate
        ? 'Publishing did not finish within 120 seconds.'
        : 'Publish did not start. Check the publish service.';
      notify();
      return;
    }

    timer = setTimeout(() => {
      void poll(baseline, deadline, progress);
    }, POLL_MS);
  }
</script>

<div class="publish">
  <div class="actions">
    {#if showButton}
      <button class="btn" type="button" disabled={busy} onclick={() => void start()}>
        {busy ? (removal ? 'Removing...' : 'Publishing...') : 'Publish to site'}
      </button>
    {/if}

    {#if busy}
      <span class="indicator" aria-hidden="true"></span>
      <span class="visually-hidden" role="status">{removal ? 'Removing the site' : 'Building the site'}</span>
    {/if}
  </div>

  {#if phase === 'failed'}
    <Banner title="Publish stopped" detail={failureDetail} />
  {/if}
</div>

<style>
  .publish {
    display: inline-flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-2);
    max-width: min(100%, 34rem);
  }

  .actions {
    display: inline-flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-2);
    max-width: 100%;
  }

  .indicator {
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    background: var(--color-accent);
    animation: pulse 1.2s ease-in-out infinite;
  }

  .publish :global(.banner) {
    width: 100%;
    margin-bottom: 0;
  }

  .visually-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 0.25;
      transform: scale(0.8);
    }
    50% {
      opacity: 1;
      transform: scale(1);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .indicator {
      animation: none;
      opacity: 1;
    }
  }
</style>
