<script lang="ts">
  // Ports dateHelpController.ts: a tap-triggered popover, closing on outside
  // click or Escape. Shared by the quick-add form and the expense search
  // panel's date-bounds tooltip (§3.9 rule 2 — components/ is for anything
  // more than one resource uses).
  import type { Snippet } from "svelte";
  import { Info } from "lucide";
  import Icon from "./Icon.svelte";

  interface Props {
    /** Accessible name for the toggle. */
    label: string;
    /** Heading above the list of accepted formats. */
    title: string;
    /**
     * Utilities appended to the panel, for the one caller that has to flip it
     * to the right edge so it does not overflow the viewport.
     */
    panelClass?: string;
    children: Snippet;
  }

  let { label, title, panelClass = "", children }: Props = $props();
  let open = $state(false);

  function toggle(event: MouseEvent): void {
    event.preventDefault();
    event.stopPropagation();
    open = !open;
  }

  function closeOnOutsideOrEscape(node: HTMLElement) {
    function onDocClick(event: MouseEvent): void {
      if (!node.contains(event.target as Node)) {
        open = false;
      }
    }
    function onKey(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        open = false;
      }
    }

    document.addEventListener("click", onDocClick);
    document.addEventListener("keydown", onKey);

    return {
      destroy(): void {
        document.removeEventListener("click", onDocClick);
        document.removeEventListener("keydown", onKey);
      },
    };
  }
</script>

<span class="relative inline-flex" use:closeOnOutsideOrEscape>
  <button
    type="button"
    class="inline-flex h-4 w-4 items-center justify-center text-muted hover:text-primary"
    onclick={toggle}
    aria-label={label}
    aria-expanded={open}
  >
    <Icon icon={Info} class="h-4 w-4" />
  </button>
  <div
    class="absolute top-[calc(100%+0.5rem)] left-0 z-100 w-max max-w-64 rounded-xs border border-line bg-surface p-3 shadow-popover {panelClass}"
    role="tooltip"
    hidden={!open}
  >
    <p class="mb-2 text-sm font-semibold text-fg">{title}</p>
    <!-- The list styling lives here rather than at each call site: both
         callers pass the same shape (a <ul> of <li>s with <code> in them) and
         had their own copy of the utilities for it. -->
    <div
      class="text-sm text-muted [&_code]:text-fg [&_li]:mb-1 [&_ul]:list-none [&_ul]:pl-4"
    >
      {@render children()}
    </div>
  </div>
</span>
