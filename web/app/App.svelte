<script lang="ts">
  import Header from "./components/Header.svelte";
  import Footer from "./components/Footer.svelte";
  import Spinner from "./components/Spinner.svelte";
  import { subscribe as subscribePending } from "./lib/pending";
  import {
    routes,
    matchRoute,
    toRoutePath,
    onPopState,
    onLinkClick,
  } from "./router";

  let path = $state(toRoutePath(window.location.pathname));
  // A resource's query string (filters, sort, pagination) can change with the
  // pathname unchanged, which leaves `path` reassigned to an identical value
  // and skips $derived recomputation — by design, since the routed component
  // should not remount over a page/sort/filter change. Passing `search`
  // through as its own prop is what lets a route react to that anyway: it
  // always changes, so Svelte always re-renders whatever reads it.
  let search = $state(window.location.search);
  // §3.7 calls this a router-level flag, but the router is the wrong owner: it
  // knows a route changed, not that the route is still fetching, so a flag it
  // set would clear before the first row arrived. lib/pending.ts counts
  // requests in flight instead and lib/api.ts drives it.
  let pending = $state(false);

  function syncLocation(nextPath: string): void {
    path = nextPath;
    search = window.location.search;
  }

  $effect(() => {
    const offPopState = onPopState(syncLocation);
    const offLinkClick = onLinkClick(syncLocation);
    const offPending = subscribePending((visible) => {
      pending = visible;
    });

    return () => {
      offPopState();
      offLinkClick();
      offPending();
    };
  });

  const match = $derived(matchRoute(routes, path));
</script>

<div class="mx-auto w-full max-w-content px-4 pt-6 pb-8 md:pt-8 md:pb-12">
  <Header />
  <Spinner visible={pending} />
  <main class="grid gap-4 md:gap-6">
    {#if match}
      {@const Route = match.component}
      <Route {...match.params} {search} />
    {:else}
      <p class="text-muted">Not found.</p>
    {/if}
  </main>
  <Footer />
</div>
