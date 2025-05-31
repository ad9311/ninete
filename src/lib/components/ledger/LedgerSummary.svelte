<script lang="ts">
	import type { Ledger } from "$lib/server/db/schema";
	import { getBalance } from "$lib/shared/ledger";
	import Amount from "$lib/components/Amount.svelte";

  export let ledger: Ledger;

  const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
  const linkRoot = ledger.type === 'budget' ? '/ledgers/budgets' : '/ledgers/accounts';
</script>

<div>
  <div class="flex justify-between items-center">
    <h2 class="text-2xl font-bold">{title}</h2>
    <a href={`${linkRoot}/${ledger.id}`}>View</a>
  </div>
  <p>Credits: <Amount value={ledger.totalCredits} type="credit" /></p>
  <p>Debits: <Amount value={ledger.totalDebits} type="debit" /></p>
  <p>Balance: <Amount value={getBalance(ledger)} type="balance" /></p>
</div>
