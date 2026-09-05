<script lang="ts">
  // Icons mount per component rather than through one global data-lucide +
  // createIcons() DOM scan: there is no app-wide lifecycle event to hang a
  // one-shot scan on (§2.3 of docs/spa-migration.md:
  // "Per-component icon rendering"). Builds the real <svg> via lucide's own
  // createElement and swaps it in for the placeholder element — DOM APIs, not
  // innerHTML, so §3.4 rule 3's {@html} ban never enters the picture. The
  // class goes on the svg itself, matching how createIcons carried the
  // original element's class over, so a caller's sizing utilities land on the
  // element that actually paints.
  import { createElement, type IconNode } from "lucide";

  interface Props {
    icon: IconNode;
    class?: string;
  }

  let { icon, class: className = "" }: Props = $props();

  function mount(node: HTMLElement) {
    const svg = createElement(icon, className ? { class: className } : {});
    // Carry the placeholder's own attributes (here, just aria-hidden) onto
    // the generated svg so the icon stays hidden from assistive tech instead
    // of reaching it unlabeled.
    for (const attr of Array.from(node.attributes)) {
      svg.setAttribute(attr.name, attr.value);
    }
    node.replaceWith(svg);

    return {
      destroy(): void {
        svg.remove();
      },
    };
  }
</script>

<i use:mount aria-hidden="true"></i>
