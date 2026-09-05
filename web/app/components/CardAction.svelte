<script lang="ts">
  // One icon action in a Card header. Renders an anchor when it navigates and
  // a button when it runs something, so the two never look different — the
  // hand-written version needed five attributes and a nested <Icon> at every
  // one of its two dozen call sites.
  //
  // `label` is both the accessible name and the hover title: the control has
  // no text of its own, so leaving either off makes it unusable in one of the
  // two ways people reach it.
  import type { IconNode } from "lucide";
  import Icon from "./Icon.svelte";

  interface Props {
    icon: IconNode;
    label: string;
    /** Set for a link. Leave unset and pass `onclick` for an action. */
    href?: string;
    onclick?: () => void;
    /** Passed through for a link the client router must not swallow. */
    rel?: string;
    disabled?: boolean;
  }

  let { icon, label, href, onclick, rel, disabled = false }: Props = $props();

  const shared =
    "text-muted hover:text-fg focus-visible:text-fg inline-flex items-center justify-center no-underline";
</script>

{#if href}
  <a {href} {rel} class={shared} aria-label={label} title={label}>
    <Icon {icon} class="h-4 w-4" />
  </a>
{:else}
  <button
    type="button"
    class={shared}
    aria-label={label}
    title={label}
    {disabled}
    {onclick}
  >
    <Icon {icon} class="h-4 w-4" />
  </button>
{/if}
