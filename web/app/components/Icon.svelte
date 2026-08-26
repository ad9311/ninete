<script lang="ts">
  // Single-icon replacement for the old data-lucide + createIcons() DOM scan
  // (web/app/lib/icons.ts), which only ran on turbo:load/turbo:render and has
  // no equivalent lifecycle event in the SPA (§2.3 of docs/spa-migration.md:
  // "Per-component icon rendering"). Builds the real <svg> via lucide's own
  // createElement and swaps it in for the placeholder element — DOM APIs, not
  // innerHTML, so §3.4 rule 3's {@html} ban never enters the picture. The
  // class goes on the svg itself, matching how createIcons carried the
  // original element's class over, so existing rules like .card-action-icon
  // apply unchanged.
  import { createElement, type IconNode } from "lucide";

  interface Props {
    icon: IconNode;
    class?: string;
  }

  let { icon, class: className = "" }: Props = $props();

  function mount(node: HTMLElement) {
    const svg = createElement(icon, className ? { class: className } : {});
    node.replaceWith(svg);

    return {
      destroy(): void {
        svg.remove();
      },
    };
  }
</script>

<i use:mount aria-hidden="true"></i>
