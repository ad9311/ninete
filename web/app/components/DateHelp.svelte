<script lang="ts">
  // Ports dateHelpController.ts: a tap-triggered popover, closing on outside
  // click or Escape. The expense search panel used to render a second one for
  // its date bounds; the quick-add form's date-format help is the only caller
  // left, so this now sits in components/ with one user rather than the two
  // §3.9 rule 2 asks for.
  import type { Snippet } from "svelte";
  import { Info } from "lucide";
  import Icon from "./Icon.svelte";

  interface Props {
    /** Accessible name for the toggle. */
    label: string;
    /** Heading above the list of accepted formats. */
    title: string;
    children: Snippet;
  }

  let { label, title, children }: Props = $props();
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
    class="absolute top-[calc(100%+0.5rem)] left-0 z-100 w-max max-w-[min(16rem,calc(100vw-2rem))] rounded-xs border border-line bg-surface p-3 shadow-popover"
    role="tooltip"
    hidden={!open}
  >
    <p class="mb-2 text-sm font-semibold text-fg">{title}</p>
    <!-- The list styling lives here rather than at the call site: the caller
         passes a <ul> of <li>s with <code> in them and used to carry its own
         copy of the utilities for it. -->
    <div
      class="text-sm text-muted [&_code]:text-fg [&_li]:mb-1 [&_ul]:list-none [&_ul]:pl-4"
    >
      {@render children()}
    </div>
  </div>
</span>
