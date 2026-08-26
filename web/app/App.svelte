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
  // No route in the Phase 1 match table loads async data yet; the flag exists
  // now so Spinner's wiring does not have to be revisited when Phase 2 adds a
  // route that does (§3.7 of docs/spa-migration.md).
  let pending = $state(false);

  $effect(() => {
    const offPopState = onPopState((next) => (path = next));
    const offLinkClick = onLinkClick((next) => (path = next));

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
      <Route {...match.params} />
    {:else}
      <p>Not found.</p>
    {/if}
  </main>
  <Footer />
</div>
