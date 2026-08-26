<script lang="ts">
  import Header from "./components/Header.svelte";
  import Footer from "./components/Footer.svelte";
  import Spinner from "./components/Spinner.svelte";
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
  // No route in the Phase 1 match table loads async data yet; the flag exists
  // now so Spinner's wiring does not have to be revisited when Phase 2 adds a
  // route that does (§3.7 of docs/spa-migration.md).
  let pending = $state(false);

  function syncLocation(nextPath: string): void {
    path = nextPath;
    search = window.location.search;
  }

  $effect(() => {
    const offPopState = onPopState(syncLocation);
    const offLinkClick = onLinkClick(syncLocation);

    return () => {
      offPopState();
      offLinkClick();
    };
  });

  const match = $derived(matchRoute(routes, path));
</script>

<div class="page-shell">
  <Header />
  <Spinner visible={pending} />
  <main class="page-main">
    {#if match}
      {@const Route = match.component}
      <Route {...match.params} {search} />
    {:else}
      <p>Not found.</p>
    {/if}
  </main>
  <Footer />
</div>
