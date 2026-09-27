<script lang="ts">
  // Personal access tokens: the credential the local MCP server sends as
  // "Authorization: Bearer" (docs/mcp.md). /api/tokens is reachable from a
  // browser session only — the server refuses it to a token — so this page is
  // the one place tokens are created and revoked.
  //
  // The plaintext arrives once, in the POST response, and is kept only in
  // component state until the page is left. Nothing the server stores can
  // reproduce it.
  import { ArrowLeft } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import LocalDate from "../../components/LocalDate.svelte";
  import { APIRequestError, del, get, post } from "../../lib/api";
  import { BASE_PATH } from "../../router";
  import type {
    APIToken,
    APITokenCreatedResponse,
    APITokenScope,
    APITokensResponse,
  } from "./types";

  let tokens = $state<APIToken[]>([]);
  let limit = $state(0);
  let expiryDays = $state<number[]>([]);
  let loaded = $state(false);
  let loadError = $state("");

  let name = $state("");
  let scope = $state<APITokenScope>("read");
  let expiresInDays = $state(90);
  let creating = $state(false);
  let createError = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  let secret = $state("");
  let copied = $state(false);
  let revokeError = $state("");
  let revoking = $state(false);

  // Expired tokens stay listed until revoked but no longer count against the
  // limit, matching logic.CountActiveAPITokens.
  const activeCount = $derived(tokens.filter((t) => !t.expired).length);
  const atLimit = $derived(limit > 0 && activeCount >= limit);

  async function load(): Promise<void> {
    const result = await get<APITokensResponse>("/tokens");
    tokens = result.data;
    limit = result.limit;
    expiryDays = result.expiry_days;
    loaded = true;
  }

  $effect(() => {
    let cancelled = false;

    load().catch((err) => {
      if (cancelled) return;
      loadError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    });

    return () => {
      cancelled = true;
    };
  });

  function expiryLabel(days: number): string {
    return days === 0 ? "Never" : `${days} days`;
  }

  async function createToken(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    creating = true;
    createError = "";
    fieldErrors = {};
    secret = "";
    copied = false;

    try {
      const result = await post<APITokenCreatedResponse>("/tokens", {
        name,
        scope,
        expires_in_days: expiresInDays,
      });
      secret = result.secret;
      name = "";
      await load();
    } catch (err) {
      if (err instanceof APIRequestError) {
        createError = err.message;
        fieldErrors = err.fields;
      } else {
        createError = "Something went wrong.";
      }
    } finally {
      creating = false;
    }
  }

  async function copySecret(): Promise<void> {
    try {
      await navigator.clipboard.writeText(secret);
      copied = true;
    } catch {
      // Clipboard access can be refused; the token is still selectable.
      copied = false;
    }
  }

  async function revokeToken(token: APIToken): Promise<void> {
    if (
      !confirm(
        `Revoke "${token.name}"? Anything using it stops working immediately.`,
      )
    ) {
      return;
    }

    revoking = true;
    revokeError = "";
    try {
      await del(`/tokens/${token.id}`);
      await load();
    } catch (err) {
      revokeError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      revoking = false;
    }
  }
</script>

<Card title="API tokens" actionsLabel="API token actions">
  {#snippet actions()}
    <CardAction
      icon={ArrowLeft}
      label="Back to account"
      href={`${BASE_PATH}/account`}
    />
  {/snippet}
  <p class="text-sm text-muted">
    A token lets a program outside the browser — such as the Ninete MCP server —
    use your account. A read token can only look; a write token can also create
    and edit. No token can delete anything.
  </p>

  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}

  {#if secret}
    <div class="grid gap-2 rounded-xs border border-line p-3">
      <p class="font-bold">Copy your new token now</p>
      <p class="text-sm text-muted">
        This is the only time it is shown. If you lose it, revoke it and create
        another.
      </p>
      <code class="break-all">{secret}</code>
      <button
        type="button"
        class="btn justify-self-end"
        onclick={() => void copySecret()}
      >
        {copied ? "Copied" : "Copy token"}
      </button>
    </div>
  {/if}

  <form onsubmit={createToken} class="grid max-w-form gap-3">
    <label>
      Name
      <input
        type="text"
        maxlength="50"
        placeholder="Laptop MCP"
        bind:value={name}
        aria-invalid={fieldErrors.name ? "true" : undefined}
      />
    </label>
    <label>
      Access
      <select bind:value={scope}>
        <option value="read">Read only</option>
        <option value="write">Read and write</option>
      </select>
    </label>
    <label>
      Expires after
      <select bind:value={expiresInDays}>
        {#each expiryDays as days (days)}
          <option value={days}>{expiryLabel(days)}</option>
        {/each}
      </select>
    </label>

    {#if createError}
      <p class="text-danger">{createError}</p>
    {/if}
    {#if loaded && atLimit}
      <p class="text-sm text-muted">
        You have {limit} active tokens, the maximum. Revoke one to create another.
      </p>
    {/if}

    <button
      type="submit"
      class="btn btn-primary justify-self-end"
      disabled={creating || !loaded || atLimit}
    >
      {creating ? "Creating..." : "Create token"}
    </button>
  </form>
</Card>

{#if loaded}
  <Card title="Your tokens" level={2}>
    {#if revokeError}
      <p class="text-danger">{revokeError}</p>
    {/if}
    {#if tokens.length === 0}
      <p class="text-sm text-muted">You have no tokens.</p>
    {:else}
      <ul class="grid gap-2">
        {#each tokens as token (token.id)}
          <li
            class="flex flex-wrap items-center justify-between gap-3 rounded-xs border border-line px-4 py-3"
          >
            <span class="grid gap-1">
              <span class="font-bold">
                {token.name}
                <span class="text-sm font-normal text-muted">
                  · {token.scope === "write" ? "read and write" : "read only"}
                </span>
              </span>
              <code class="text-sm text-muted">{token.prefix}…</code>
              <span class="text-sm text-muted">
                Created <LocalDate value={token.created_at} datetime />
                ·
                {#if token.last_used_at}
                  Last used <LocalDate value={token.last_used_at} datetime />
                {:else}
                  Never used
                {/if}
                ·
                {#if token.expires_at === null}
                  Never expires
                {:else if token.expired}
                  <span class="text-danger">
                    Expired <LocalDate value={token.expires_at} datetime />
                  </span>
                {:else}
                  Expires <LocalDate value={token.expires_at} datetime />
                {/if}
              </span>
            </span>
            <button
              type="button"
              class="btn btn-danger"
              disabled={revoking}
              onclick={() => void revokeToken(token)}
            >
              Revoke
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </Card>
{/if}
