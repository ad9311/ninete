<script lang="ts">
  // A sortable column heading. The three tables that have one rendered it by
  // hand against `.sort-link`/`.sort-indicator`, two class names that were
  // never defined in the stylesheet at all — the arrow and the label were
  // unstyled everywhere and nobody noticed, which is the argument for having
  // one of these.
  interface Props {
    label: string;
    /** Where clicking sorts by this column, from the caller's sortHref(). */
    href: string;
    /** True when the table is currently sorted by this column. */
    active: boolean;
    /** The current order, only meaningful while `active`. */
    order: string;
  }

  let { label, href, active, order }: Props = $props();
</script>

<th
  aria-sort={active ? (order === "ASC" ? "ascending" : "descending") : "none"}
>
  <a
    {href}
    class="inline-flex items-center gap-1 text-muted no-underline hover:text-primary"
  >
    {label}
    {#if active}
      <span aria-hidden="true">{order === "ASC" ? "▲" : "▼"}</span>
    {/if}
  </a>
</th>
