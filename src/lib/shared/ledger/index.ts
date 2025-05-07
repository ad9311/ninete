import type { Ledger } from '$lib/server/db/schema';

export function getBalance(ledger: Ledger): number {
	const totalCredits = Number(ledger.totalCredits);
	const totalDebits = Number(ledger.totalDebits);

	return totalCredits - totalDebits;
}
