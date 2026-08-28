<script lang="ts">
  // Ports login/index.html. On success the client does a full page load
  // rather than staying inside the SPA (§5, Phase 6 note in
  // docs/spa-migration.md): RenewToken/session state is a boundary every
  // piece of client-held state needs reset against, and a hard navigation is
  // the simplest way to get that. BASE_PATH itself, not "/", so the login
  // that started inside the SPA lands back inside it.
  import { APIRequestError, post } from "../../lib/api";
  import { BASE_PATH } from "../../router";

  let email = $state("");
  let password = $state("");
  let error = $state("");
  let pending = $state(false);

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    pending = true;
    try {
      await post("/login", { email, password });
      window.location.assign(BASE_PATH);
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
      pending = false;
    }
  }
</script>

<section class="auth-page">
  <div class="auth-card">
    <h1>LOGIN</h1>
    {#if error}
      <p class="form-error-text">{error}</p>
    {/if}
    <form onsubmit={handleSubmit} class="form-stack">
      <label>
        Email
        <input
          type="email"
          name="email"
          autocomplete="email"
          bind:value={email}
        />
      </label>
      <label>
        Password
        <input
          type="password"
          name="password"
          autocomplete="current-password"
          bind:value={password}
        />
      </label>
      <button type="submit" class="btn-primary form-submit" disabled={pending}>
        {pending ? "Logging in..." : "Login"}
      </button>
    </form>
    <p class="auth-switch">
      Need an account? <a href={`${BASE_PATH}/register`}>Register</a>
    </p>
  </div>
</section>
