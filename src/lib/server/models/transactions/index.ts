import { createInsertSchema } from 'drizzle-zod';
import { z } from 'zod';
import { db } from '$lib/server/db';
import { transactionsTable, type Transaction } from '$lib/server/db/schema';
import { onTransactionCommit } from '../ledger/budget';
import { eq } from 'drizzle-orm';

export const transactionCreateSchema = createInsertSchema(transactionsTable, {
	ledgerId: (schema) => schema.int().positive({ message: 'Budget ID must be a positive integer' }),
	amount: (schema) =>
		schema.refine(
			(value) => {
				const num = Number(value);
				return !isNaN(num) && num > 0;
			},
			{ message: 'Amount must be a positive number' }
		),
	description: (schema) =>
		schema
			.nonempty({ message: 'Description is required' })
			.max(255, { message: 'Description must be less than 255 characters' }),
	date: (schema) =>
		schema.refine(
			(input) => {
				const today = new Date();
				return input <= today;
			},
			{ message: 'Date cannot be in the future' }
		),
	category: (schema) =>
		schema.exclude(['payable', 'receivable'], { message: 'Category is required' }),
	type: (schema) => schema
});

export const transactionUpdateSchema = transactionCreateSchema
	.omit({
		ledgerId: true
	})
	.partial({
		type: true
	});

export type TransactionCreateData = z.infer<typeof transactionCreateSchema>;
export type TransactionUpdateData = z.infer<typeof transactionUpdateSchema>;

export async function findTransaction(transactionId: number): Promise<Transaction | undefined> {
	return await db.query.transactionsTable.findFirst({
		where: eq(transactionsTable.id, transactionId)
	});
}

export async function createTransaction(params: TransactionCreateData): Promise<Transaction> {
	const validated = transactionCreateSchema.parse(params);

	return db.transaction(async (tx) => {
		const result = await tx.insert(transactionsTable).values(validated).returning();
		await onTransactionCommit(tx, validated.ledgerId, {
			previousAmount: 0,
			newAmount: validated.amount,
			commitColumn: validated.type === 'credit' ? 'totalCredits' : 'totalDebits'
		});

		return result[0];
	});
}

export async function updateTransaction(
	budgetId: number,
	transactionId: number,
	params: TransactionUpdateData
): Promise<Transaction> {
	const validated = transactionUpdateSchema.parse(params);

	const transaction = await findTransaction(transactionId);

	if (params.type && params.type !== validated.type) {
		throw new Error('Type cannot be updated');
	}

	if (!transaction) {
		throw new Error('Transaction not found');
	}

	return db.transaction(async (tx) => {
		const result = await tx
			.update(transactionsTable)
			.set(validated)
			.where(eq(transactionsTable.id, transactionId))
			.returning();

		await onTransactionCommit(tx, budgetId, {
			previousAmount: transaction.amount,
			newAmount: validated.amount,
			commitColumn: transaction.type === 'credit' ? 'totalCredits' : 'totalDebits'
		});

		return result[0];
	});
}

export async function deleteTransaction(
	budgetId: number,
	transactionId: number
): Promise<Transaction> {
	const transaction = await findTransaction(transactionId);

	if (!transaction) {
		throw new Error('Transaction not found');
	}

	return db.transaction(async (tx) => {
		const result = await tx
			.delete(transactionsTable)
			.where(eq(transactionsTable.id, transactionId))
			.returning();

		await onTransactionCommit(tx, budgetId, {
			previousAmount: transaction.amount,
			newAmount: 0,
			commitColumn: transaction.type === 'credit' ? 'totalCredits' : 'totalDebits'
		});

		return result[0];
	});
}
