<script lang="ts">
  const label = 'Continue with GitHub';
</script>

<main class="login">
  <a class="btn btn-primary sign-in" href="/auth/github">{label}</a>
</main>

<style>
  .login {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100svh;
    padding: var(--gutter);
  }

  .sign-in {
    padding: 0.75rem 1.5rem;
    font-size: var(--text-lg);
  }
</style>
