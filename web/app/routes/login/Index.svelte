<script lang="ts">
  // Ports login/index.html. On success the client does a full page load
  // rather than staying inside the SPA (§5, Phase 6 note in
  // docs/spa-migration.md): RenewToken/session state is a boundary every
  // piece of client-held state needs reset against, and a hard navigation is
  // the simplest way to get that. `BASE_PATH || "/"` rather than BASE_PATH
  // alone: BASE_PATH is "" since Phase 7, and location.assign("") resolves
  // against the current document, which reloads /login instead of navigating
  // to the dashboard. The fallback keeps the seam if the SPA is ever staged
  // under a prefix again.
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
      window.location.assign(BASE_PATH || "/");
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
      pending = false;
    }
  }
</script>

<section class="grid place-items-center md:min-h-[calc(100vh-8rem)]">
  <div
    class="grid w-full max-w-auth gap-4 rounded-xs border border-line bg-surface p-4 shadow-panel md:p-6"
  >
    <h1 class="text-center">LOGIN</h1>
    {#if error}
      <p class="text-danger">{error}</p>
    {/if}
    <form onsubmit={handleSubmit} class="grid gap-3">
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
      <button
        type="submit"
        class="btn btn-primary mt-3 justify-self-end"
        disabled={pending}
      >
        {pending ? "Logging in..." : "Login"}
      </button>
    </form>
    <p class="text-center text-muted">
      Need an account? <a href={`${BASE_PATH}/register`}>Register</a>
    </p>
  </div>
</section>
