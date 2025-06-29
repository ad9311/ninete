import { db } from '$lib/server/db';
import { and, eq } from 'drizzle-orm';
import {
	ledgersTable,
	transactionsTable,
	type Ledger,
	type Transaction
} from '$lib/server/db/schema';
import type { NewBudgetParams } from '$lib/server/models/ledger/budget';
import type { LEDGER_TYPE } from '$lib/shared';
import type { NewLoanParams } from './loans';

export const LEDGER_ERRORS = {
	userId: 'User ID must be a positive integer',
	titleNonEmpty: 'Title cannot be empty',
	titleMax: 'Title must less than 50 characters',
	description: 'Description must be less than 100 characters',
	year: 'Year must be a positive integer',
	month: 'Month must be a positive integer'
} as const;

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

export async function createLedger(params: NewBudgetParams | NewLoanParams): Promise<Ledger> {
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
	type: LEDGER_TYPE
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
	type: LEDGER_TYPE
): Promise<Ledger | undefined> {
	return await db.query.ledgersTable.findFirst({
		where: and(eq(ledgersTable.id, ledgerId), eq(ledgersTable.type, type))
	});
}

export async function findLedgers(userId: number, type: LEDGER_TYPE): Promise<Ledger[]> {
	return await db
		.select()
		.from(ledgersTable)
		.where(and(eq(ledgersTable.userId, userId), eq(ledgersTable.type, type)));
}
