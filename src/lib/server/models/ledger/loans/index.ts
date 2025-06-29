import { ledgersTable, type Ledger } from '$lib/server/db/schema';
import { createInsertSchema } from 'drizzle-zod';
import { z } from 'zod';
import { createLedger, LEDGER_ERRORS } from '..';

export const loanCreateSchema = createInsertSchema(ledgersTable, {
	userId: (schema) => schema.int().positive({ message: LEDGER_ERRORS.userId }),
	title: (schema) =>
		schema
			.nonempty({ message: LEDGER_ERRORS.titleNonEmpty })
			.max(50, { message: LEDGER_ERRORS.titleMax }),
	description: (schema) => schema.max(100, { message: LEDGER_ERRORS.description }),
	year: (schema) => schema.int().positive({ message: LEDGER_ERRORS.year }),
	month: (schema) => schema.int().positive({ message: LEDGER_ERRORS.month }),
	type: (schema) =>
		schema.exclude(['budget', 'savings'], {
			message: 'Type must be of loan'
		}),
	status: (schema) => schema
});

export const loanNewSchema = loanCreateSchema.omit({ year: true, month: true }).extend({
	date: z.date()
});

export type CreateLoan = z.infer<typeof loanCreateSchema>;
export type NewLoanParams = z.infer<typeof loanNewSchema>;

export async function createLoan(params: NewLoanParams): Promise<Ledger> {
	return await createLedger(params);
}
