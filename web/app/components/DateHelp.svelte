<script lang="ts">
  // Ports dateHelpController.ts: a tap-triggered popover, closing on outside
  // click or Escape. Shared by the quick-add form and the expense search
  // panel's date-bounds tooltip (§3.9 rule 2 — components/ is for anything
  // more than one resource uses).
  import type { Snippet } from "svelte";
  import { Info } from "lucide";
  import Icon from "./Icon.svelte";

  interface Props {
    label: string;
    panelClass?: string;
    children: Snippet;
  }

  let { label, panelClass = "", children }: Props = $props();
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

<span class="date-help" use:closeOnOutsideOrEscape>
  <button
    type="button"
    class="date-help-toggle"
    onclick={toggle}
    aria-label={label}
    aria-expanded={open}
  >
    <Icon icon={Info} class="date-help-icon" />
  </button>
  <div class="date-help-panel {panelClass}" role="tooltip" hidden={!open}>
    {@render children()}
  </div>
</span>
