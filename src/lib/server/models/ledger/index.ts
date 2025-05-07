import { db } from '$lib/server/db';
import { and, eq } from 'drizzle-orm';
import { transactionsTable, type Transaction } from '$lib/server/db/schema';

export async function findLedgetCredtis(ledgerId: number): Promise<Transaction[]> {
	const credits = await db.query.transactionsTable.findMany({
		where: and(eq(transactionsTable.ledgerId, ledgerId), eq(transactionsTable.type, 'credit'))
	});

	return credits;
}

export async function findLedgetDebits(ledgerId: number): Promise<Transaction[]> {
	const debits = await db.query.transactionsTable.findMany({
		where: and(eq(transactionsTable.ledgerId, ledgerId), eq(transactionsTable.type, 'debit'))
	});

	return debits;
}
