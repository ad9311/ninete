import { db } from '$lib/server/db';
import { and, eq } from 'drizzle-orm';
import {
	ledgersTable,
	transactionsTable,
	type Ledger,
	type Transaction
} from '$lib/server/db/schema';
import type { NewBudgetParams } from '$lib/server/models/ledger/budget';
import type { LEDGER_TYPES } from '$lib/shared';

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

export async function createLedger(params: NewBudgetParams): Promise<Ledger> {
	const { date, ...rest } = params;

	const year = date.getFullYear();
	const month = date.getMonth() + 1;

	const createParams = {
		...rest,
		year,
		month
	};

	const result = await db.insert(ledgersTable).values(createParams).returning();

	return result[0];
}

export async function findLedger(
	userId: number,
	ledgerId: number,
	type: (typeof LEDGER_TYPES)[number]
): Promise<Ledger | undefined> {
	return await db.query.ledgersTable.findFirst({
		where: and(
			eq(ledgersTable.userId, userId),
			eq(ledgersTable.id, ledgerId),
			eq(ledgersTable.type, type)
		)
	});
}

export async function findLedgerById(
	ledgerId: number,
	type: (typeof LEDGER_TYPES)[number]
): Promise<Ledger | undefined> {
	return await db.query.ledgersTable.findFirst({
		where: and(eq(ledgersTable.id, ledgerId), eq(ledgersTable.type, type))
	});
}

export async function findLedgers(
	userId: number,
	type: (typeof LEDGER_TYPES)[number]
): Promise<Ledger[]> {
	return await db
		.select()
		.from(ledgersTable)
		.where(and(eq(ledgersTable.userId, userId), eq(ledgersTable.type, type)));
}
