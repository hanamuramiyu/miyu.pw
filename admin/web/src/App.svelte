<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiError, api, errorNotice, setCSRFToken, type Notice } from './lib/api';
  import { navigate, normalizePath } from './lib/router';
  import type { Me } from './lib/types';
  import Banner from './lib/components/Banner.svelte';
  import Login from './routes/Login.svelte';
  import WorkEditor from './routes/WorkEditor.svelte';
  import WorkList from './routes/WorkList.svelte';

  type Route =
    | { name: 'list' }
    | { name: 'editor'; id: string | null }
    | { name: 'notfound' };

  function parseRoute(path: string): Route {
    if (path === '/') return { name: 'list' };
    if (path === '/works/new') return { name: 'editor', id: null };

    const match = /^\/works\/([^/]+)$/.exec(path);
    if (match) return { name: 'editor', id: decodeURIComponent(match[1]) };

    return { name: 'notfound' };
  }

  let path = $state(normalizePath(window.location.pathname));
  let me = $state<Me | null>(null);
  let ready = $state(false);
  let bootError = $state<Notice | null>(null);

  const route = $derived(parseRoute(path));
  const routeName = $derived(route.name);
  const editorId = $derived(route.name === 'editor' ? route.id : null);

  onMount(() => {
    const onPop = () => {
      path = normalizePath(window.location.pathname);
    };
    window.addEventListener('popstate', onPop);

    api
      .me()
      .then((value) => {
        setCSRFToken(value.csrfToken);
        me = value;
      })
      .catch((error: unknown) => {
        if (error instanceof ApiError && error.status === 401) return;
        bootError = errorNotice(error);
      })
      .finally(() => {
        ready = true;
      });

    return () => window.removeEventListener('popstate', onPop);
  });

  function go(to: string, replace = false): void {
    navigate(to, replace);
  }

  function unauthorized(): void {
    setCSRFToken('');
    me = null;
    go('/');
  }

  function signedOut(): void {
    setCSRFToken('');
    me = null;
    go('/');
  }
</script>

{#if !ready}
  <p class="page muted">Loading...</p>
{:else if me === null && bootError !== null}
  <div class="page">
    <Banner title={bootError.title} detail={bootError.detail} />
    <button class="btn" type="button" onclick={() => window.location.reload()}>Retry</button>
  </div>
{:else if me === null}
  <Login />
{:else if routeName === 'list'}
  <WorkList {me} onNavigate={go} onUnauthorized={unauthorized} onSignedOut={signedOut} />
{:else if routeName === 'editor'}
  {#key editorId}
    <WorkEditor id={editorId} onNavigate={go} onUnauthorized={unauthorized} />
  {/key}
{:else}
  <div class="page">
    <div class="topbar">
      <h1>Page not found</h1>
    </div>
    <p class="muted">This page does not exist.</p>
    <a href="/" onclick={(event) => { event.preventDefault(); go('/'); }}>Back to works</a>
  </div>
{/if}
