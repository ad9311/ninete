<script lang="ts">
  // Theme and nav behavior for the shell (docs/spa-migration.md §2.2):
  // localStorage key `theme`, theme-light/theme-dark classes on <html>, and an
  // outside click closing the dropdown. The inline anti-FOUC script in the
  // shell already set the class before this mounts, so applying it again here
  // is a re-assertion, not the first write.
  import { get, csrfToken } from "../lib/api";
  import { BASE_PATH } from "../router";

  type Theme = "light" | "dark" | "auto";
  const THEME_KEY = "theme";

  interface Session {
    id: number;
    username: string;
    email: string;
  }

  function readStoredTheme(): Theme {
    try {
      const stored = localStorage.getItem(THEME_KEY);
      if (stored === "light" || stored === "dark" || stored === "auto") {
        return stored;
      }
    } catch {
      // Storage can be unavailable (private mode, disabled cookies).
    }

    return "auto";
  }

  const media = window.matchMedia("(prefers-color-scheme: dark)");

  let theme = $state<Theme>(readStoredTheme());
  let session = $state<Session | null>(null);
  let navOpen = $state(false);

  function applyTheme(value: Theme): void {
    const dark = value === "dark" || (value === "auto" && media.matches);
    document.documentElement.className = dark ? "theme-dark" : "theme-light";
  }

  $effect(() => {
    applyTheme(theme);
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch {
      // Persistence is best-effort, same as readStoredTheme above.
    }
  });

  $effect(() => {
    const onSystemChange = () => {
      if (theme === "auto") applyTheme("auto");
    };
    media.addEventListener("change", onSystemChange);

    return () => media.removeEventListener("change", onSystemChange);
  });

  $effect(() => {
    let cancelled = false;

    // skipAuthRedirect only on the two routes where a 401 here is expected
    // (Phase 6): a guest landing on /login or /register would otherwise get
    // bounced straight back by their own session check. Header
    // mounts once for the life of the SPA shell (it lives outside the
    // route-switching block in App.svelte), so this reads the URL the shell
    // was loaded with, which is exactly when this effect runs. Everywhere
    // else a 401 here still has to drive the redirect: some routes (e.g.
    // routes/account/Index.svelte) make no /api/* call of their own, so this
    // probe is the only thing that would ever notice an expired session on
    // them.
    const isGuestPage =
      window.location.pathname === `${BASE_PATH}/login` ||
      window.location.pathname === `${BASE_PATH}/register`;

    get<Session>("/session", { skipAuthRedirect: isGuestPage })
      .then((result) => {
        if (!cancelled) session = result;
      })
      .catch(() => {
        // Nothing to do for any of them: a redirect already fired for a real
        // 401 outside the guest pages; a guest's expected 401 is
        // deliberately not redirected; and any other failure just leaves the
        // chrome without a signed-in user.
      });

    return () => {
      cancelled = true;
    };
  });

  // One table rather than four copies of the same anchor, with `divided`
  // marking where the rule goes: the dropdown link styling is a long utility
  // string, and hand-written copies of it are chances to drift.
  const NAV_LINKS = [
    { href: "/expenses", label: "Expenses" },
    { href: "/recurrent-expenses", label: "Recurrent Expenses" },
    { href: "/expenses/budgets", label: "Expense Budgets" },
    { href: "/account", label: "Account", divided: true },
  ];

  const navLinkClass =
    "text-fg hover:bg-surface-hover hover:text-primary block px-4 py-2 text-sm no-underline";

  function toggleNav(): void {
    navOpen = !navOpen;
  }

  function closeOnOutsideClick(node: HTMLElement) {
    function handler(event: MouseEvent): void {
      if (!node.contains(event.target as Node)) {
        navOpen = false;
      }
    }
    document.addEventListener("click", handler);

    return {
      destroy(): void {
        document.removeEventListener("click", handler);
      },
    };
  }
</script>

<header
  class="mb-6 flex items-center gap-3 rounded-xs border border-line bg-surface px-4 py-3"
>
  <!-- Literal "/" rather than BASE_PATH: an empty href attribute (BASE_PATH
       since Phase 7 of docs/spa-migration.md) drops the anchor's implicit
       link role. -->
  <a
    href="/"
    class="mr-auto text-xl font-bold tracking-wider text-fg no-underline"
  >
    NINETE
  </a>
  <label>
    <span class="sr-only">Theme</span>
    <!-- The compact chrome controls override the app-wide control styling
         rather than opting out of it: a utility sits in a later layer than the
         base rule in app.css, so it simply wins. -->
    <select
      class="min-h-0 w-auto cursor-pointer rounded-xs border border-line bg-neutral px-3 py-2 text-sm font-medium text-fg hover:bg-neutral-hover"
      bind:value={theme}
    >
      <option value="auto">Automatic</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
    </select>
  </label>
  {#if session}
    <nav class="relative" use:closeOnOutsideClick>
      <button
        type="button"
        class="inline-flex items-center gap-1 rounded-xs border border-line bg-neutral px-3 py-2 text-sm font-medium text-fg hover:bg-neutral-hover"
        aria-haspopup="true"
        onclick={toggleNav}
      >
        {session.username}
        <span class="text-[0.7em] leading-none">▾</span>
      </button>
      <ul
        class="absolute top-[calc(100%+0.5rem)] right-0 z-100 min-w-48 rounded-xs border border-line bg-surface py-2 shadow-popover"
        class:hidden={!navOpen}
      >
        {#each NAV_LINKS as link (link.href)}
          {#if link.divided}
            <li class="my-2 h-px bg-line"></li>
          {/if}
          <li>
            <a href={`${BASE_PATH}${link.href}`} class={navLinkClass}>
              {link.label}
            </a>
          </li>
        {/each}
        <li>
          <form action="/logout" method="post" class="block">
            <input type="hidden" name="csrf_token" value={csrfToken()} />
            <button
              type="submit"
              class="block w-full px-4 py-2 text-left text-sm text-danger hover:bg-surface-hover"
            >
              Logout
            </button>
          </form>
        </li>
      </ul>
    </nav>
  {/if}
</header>
