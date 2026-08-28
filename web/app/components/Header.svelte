<script lang="ts">
  // Ports web/static/js/controllers/themeController.ts and navController.ts
  // (docs/spa-migration.md §2.2): same localStorage key, same
  // theme-light/theme-dark classes on <html>, same outside-click-closes
  // dropdown behavior. The inline anti-FOUC script in the shell already set
  // the class before this mounts; applying it again here matches what the
  // Stimulus controller's connect() did too.
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

    // skipAuthRedirect: this probe is expected to 401 for a guest, including
    // on the now-reachable /app/login and /app/register (Phase 6) — without
    // it, a guest landing on either gets bounced straight back by their own
    // session check.
    get<Session>("/session", { skipAuthRedirect: true })
      .then((result) => {
        if (!cancelled) session = result;
      })
      .catch(() => {
        // A 401 already redirected to /login (lib/api.ts); any other failure
        // just leaves the chrome without a signed-in user.
      });

    return () => {
      cancelled = true;
    };
  });

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

<header class="site-header">
  <a href={BASE_PATH} class="site-brand">NINETE</a>
  <label class="theme-switch">
    <span class="sr-only">Theme</span>
    <select bind:value={theme}>
      <option value="auto">Automatic</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
    </select>
  </label>
  {#if session}
    <nav class="site-nav" use:closeOnOutsideClick>
      <button
        type="button"
        class="site-nav-toggle"
        aria-haspopup="true"
        onclick={toggleNav}
      >
        {session.username} <span class="site-nav-caret">▾</span>
      </button>
      <ul class="site-nav-dropdown" class:open={navOpen}>
        <li><a href={`${BASE_PATH}/expenses`}>Expenses</a></li>
        <li>
          <a href={`${BASE_PATH}/recurrent-expenses`}>Recurrent Expenses</a>
        </li>
        <li><a href={`${BASE_PATH}/expenses/budgets`}>Expense Budgets</a></li>
        <li class="site-nav-divider"></li>
        <li><a href={`${BASE_PATH}/account`}>Account</a></li>
        <li>
          <form action="/logout" method="post" class="site-nav-logout">
            <input type="hidden" name="csrf_token" value={csrfToken()} />
            <button type="submit" class="site-nav-logout-btn">Logout</button>
          </form>
        </li>
      </ul>
    </nav>
  {/if}
</header>
