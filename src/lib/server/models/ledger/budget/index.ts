import { createInsertSchema } from 'drizzle-zod';
import { ledgersTable, type Ledger } from '$lib/server/db/schema';
import { z } from 'zod';
import { db, type DBTransaction } from '$lib/server/db';
import { and, eq } from 'drizzle-orm';
import { createLedger, LEDGER_ERRORS } from '$lib/server/models/ledger';
import { updateUpdatedAt } from '$lib/server/models';

type TransactionCommitParams = {
	previousAmount: number | string;
	newAmount: number | string;
	commitColumn: 'totalCredits' | 'totalDebits';
};

export const budgetCreateSchema = createInsertSchema(ledgersTable, {
	userId: (schema) => schema.int().positive({ message: LEDGER_ERRORS.userId }),
	year: (schema) => schema.int().positive({ message: LEDGER_ERRORS.year }),
	month: (schema) => schema.int().positive({ message: LEDGER_ERRORS.month }),
	type: (schema) =>
		schema.exclude(['savings', 'loan'], {
			message: 'Type must be of type budget'
		}),
	status: (schema) =>
		schema.exclude(['pending', 'paid', 'overdue', 'cancelled'], {
			message: 'Status must be of status n/a'
		})
});

export const newBudgetSchema = budgetCreateSchema.omit({ year: true, month: true }).extend({
	date: z.date()
});

export type CreateBudgetParams = z.infer<typeof budgetCreateSchema>;
export type NewBudgetParams = z.infer<typeof newBudgetSchema>;

export async function createBudget(params: NewBudgetParams): Promise<Ledger> {
	return await createLedger(params);
}

export async function findCurrentBudget(userId: number): Promise<Ledger | undefined> {
	const currentDate = new Date();
	const year = currentDate.getFullYear();
	const month = currentDate.getMonth() + 1;
	const budget = await db.query.ledgersTable.findFirst({
		where: and(
			eq(ledgersTable.userId, userId),
			eq(ledgersTable.type, 'budget'),
			eq(ledgersTable.year, year),
			eq(ledgersTable.month, month)
		)
	});

	return budget;
}

export async function findOrCreateBudget(userId: number): Promise<Ledger> {
	const budget = await findCurrentBudget(userId);

	if (budget) {
		return budget;
	}

	const currentDate = new Date();
	const date = new Date(currentDate.getFullYear(), currentDate.getMonth(), 1);
	const params: NewBudgetParams = {
		userId,
		date,
		type: 'budget',
		status: 'n/a'
	};

	return createBudget(params);
}

export async function onTransactionCommit(
	tx: DBTransaction,
	budget: Ledger,
	params: TransactionCommitParams
): Promise<Ledger> {
	if (!budget) {
		throw new Error('Budget not found');
	}

	const delta = Number(params.newAmount) - Number(params.previousAmount);
	const newTotal = Number(budget[params.commitColumn]) + delta;

	const result = await tx
		.update(ledgersTable)
		.set({ [params.commitColumn]: newTotal, ...updateUpdatedAt() })
		.where(eq(ledgersTable.id, budget.id))
		.returning();

	return result[0];
}
