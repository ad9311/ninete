<script lang="ts">
  // Ports register/index.html — see routes/login/Index.svelte for why success
  // navigates with a full page load instead of staying client-side.
  import { APIRequestError, post } from "../../lib/api";
  import { BASE_PATH } from "../../router";

  let username = $state("");
  let email = $state("");
  let password = $state("");
  let passwordConfirmation = $state("");
  let invitationCode = $state("");
  let error = $state("");
  let pending = $state(false);

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    pending = true;
    try {
      await post("/register", {
        username,
        email,
        password,
        password_confirmation: passwordConfirmation,
        invitation_code: invitationCode,
      });
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
    <h1 class="text-center">REGISTER</h1>
    {#if error}
      <p class="text-danger">{error}</p>
    {/if}
    <form onsubmit={handleSubmit} class="grid gap-3">
      <label>
        Username
        <input
          type="text"
          name="username"
          autocomplete="username"
          bind:value={username}
        />
      </label>
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
          autocomplete="new-password"
          bind:value={password}
        />
      </label>
      <label>
        Password confirmation
        <input
          type="password"
          name="passwordConfirmation"
          autocomplete="new-password"
          bind:value={passwordConfirmation}
        />
      </label>
      <label>
        Invitation code
        <input type="text" name="invitationCode" bind:value={invitationCode} />
      </label>
      <button
        type="submit"
        class="btn btn-primary mt-3 justify-self-end"
        disabled={pending}
      >
        {pending ? "Registering..." : "Register"}
      </button>
    </form>
    <p class="text-center text-muted">
      Already have an account? <a href={`${BASE_PATH}/login`}>Login</a>
    </p>
  </div>
</section>
