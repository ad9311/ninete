<script lang="ts">
  // Ports account/index.html. Pure navigation — no data to fetch, so unlike
  // every other route so far this one has no $effect at all.
  import { ChevronRight } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { BASE_PATH } from "../../router";
</script>

<section class="card" aria-labelledby="account-card-title">
  <header class="card-header">
    <h1 id="account-card-title" class="card-title">Account</h1>
  </header>
  <p class="card-empty">Manage your data and account settings.</p>
  <nav class="settings-menu" aria-label="Account settings">
    <a class="settings-menu-item" href={`${BASE_PATH}/account/exports`}>
      <span class="settings-menu-text">
        <span class="settings-menu-title">Exports</span>
        <span class="settings-menu-hint">
          Download your data as JSON files.
        </span>
      </span>
      <Icon icon={ChevronRight} class="settings-menu-icon" />
    </a>
    <a class="settings-menu-item" href={`${BASE_PATH}/account/delete-data`}>
      <span class="settings-menu-text">
        <span class="settings-menu-title">Delete data</span>
        <span class="settings-menu-hint">
          Permanently remove records from your account.
        </span>
      </span>
      <Icon icon={ChevronRight} class="settings-menu-icon" />
    </a>
  </nav>
</section>
