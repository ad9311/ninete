import { db } from '$lib/server/db';
import { ledgersTable, type Ledger } from '$lib/server/db/schema';
import { createInsertSchema } from 'drizzle-zod';
import { z } from 'zod';

export const payableReceivableCreateSchema = createInsertSchema(ledgersTable, {
	userId: (schema) => schema.int().positive({ message: 'User ID must be a positive integer' }),
	title: (schema) => schema.nonempty().max(50, { message: 'Title must less than 50 characters' }),
	description: (schema) =>
		schema.max(100, { message: 'Description must be less than 100 characters' }),
	year: (schema) => schema.int().positive({ message: 'Year must be a positive integer' }),
	month: (schema) => schema.int().positive({ message: 'Month must be a positive integer' }),
	type: (schema) =>
		schema.exclude(['budget', 'savings'], {
			message: 'Type must be of type payable or receivable'
		}),
	status: (schema) => schema
});

export const newPayableReceivableSchema = payableReceivableCreateSchema
	.omit({ year: true, month: true })
	.extend({
		date: z.date()
	});

export type PayableReceivableCreateData = z.infer<typeof payableReceivableCreateSchema>;
export type NewPayableReceivable = z.infer<typeof newPayableReceivableSchema>;

export async function createPayableReceivable(params: NewPayableReceivable): Promise<Ledger> {
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
