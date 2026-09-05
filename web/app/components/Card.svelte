<script lang="ts">
  // The surface every route renders into: a titled panel with an optional row
  // of icon actions. Twelve routes drew this by hand before the Tailwind
  // migration, each repeating the section/header/heading markup and its own
  // `aria-labelledby` wiring, which is exactly how the copies drifted. The
  // heading id is generated here so a caller cannot forget it.
  import type { Snippet } from "svelte";

  interface Props {
    title: string;
    /**
     * Heading level. 1 for the card that titles the page, 2 for a card inside
     * a grid of them — a page must not have two h1s.
     */
    level?: 1 | 2;
    /** Icon actions for the header's right-hand side, usually <CardAction>s. */
    actions?: Snippet;
    /** aria-label for the actions nav. Required whenever `actions` is given. */
    actionsLabel?: string;
    children: Snippet;
  }

  let {
    title,
    level = 1,
    actions,
    actionsLabel = "Actions",
    children,
  }: Props = $props();

  const titleId = $props.id();
</script>

<section
  class="grid content-start gap-3 rounded-xs border border-line bg-surface p-6 shadow-card"
  aria-labelledby={titleId}
>
  <header class="flex items-center justify-between gap-3">
    {#if level === 1}
      <h1 id={titleId} class="text-xl font-bold text-fg">{title}</h1>
    {:else}
      <h2 id={titleId} class="text-xl font-bold text-fg">{title}</h2>
    {/if}
    {#if actions}
      <nav class="flex gap-3" aria-label={actionsLabel}>
        {@render actions()}
      </nav>
    {/if}
  </header>
  {@render children()}
</section>
