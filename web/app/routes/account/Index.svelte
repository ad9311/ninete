<script lang="ts">
  // Ports account/index.html. Pure navigation — no data to fetch, so unlike
  // every other route so far this one has no $effect at all.
  import { ChevronRight } from "lucide";
  import Card from "../../components/Card.svelte";
  import Icon from "../../components/Icon.svelte";
  import { BASE_PATH } from "../../router";

  const SETTINGS = [
    {
      href: "/account/exports",
      title: "Exports",
      hint: "Download your data as JSON files.",
    },
    {
      href: "/account/delete-data",
      title: "Delete data",
      hint: "Permanently remove records from your account.",
    },
  ];
</script>

<Card title="Account">
  <p class="text-sm text-muted">Manage your data and account settings.</p>
  <nav class="grid gap-2" aria-label="Account settings">
    {#each SETTINGS as item (item.href)}
      <a
        class="flex items-center justify-between gap-3 rounded-xs border border-line px-4 py-3 text-fg no-underline hover:border-muted hover:bg-surface-muted focus-visible:border-muted focus-visible:bg-surface-muted"
        href={`${BASE_PATH}${item.href}`}
      >
        <span class="grid gap-1">
          <span class="font-bold">{item.title}</span>
          <span class="text-sm text-muted">{item.hint}</span>
        </span>
        <Icon icon={ChevronRight} class="h-4 w-4 flex-none text-muted" />
      </a>
    {/each}
  </nav>
</Card>
