import {
	TRANSACTION_CATEGORIES,
	LEDGER_STATUS,
	LEDGER_TYPES,
	TRANSACTION_TYPES
} from '../../shared';
import {
	boolean,
	integer,
	numeric,
	pgEnum,
	pgTable,
	serial,
	text,
	timestamp,
	uniqueIndex,
	type PgTimestampConfig
} from 'drizzle-orm/pg-core';
import { sql } from 'drizzle-orm';

const timestamps = {
	createdAt: timestamp('created_at', { withTimezone: true, mode: 'date' }).notNull().defaultNow(),
	updatedAt: timestamp('updated_at', { withTimezone: true, mode: 'date' }).notNull().defaultNow()
};

const numericParams = { precision: 10, scale: 2 };
const dateParams: PgTimestampConfig<'string' | 'date'> = { withTimezone: true, mode: 'date' };

export const ledgerStatusEnum = pgEnum('ledger_status', LEDGER_STATUS);
export const ledgerTypeEnum = pgEnum('ledger_type', LEDGER_TYPES);
export const transactionCategoryEnum = pgEnum('transaction_category', TRANSACTION_CATEGORIES);
export const transactionTypeEnum = pgEnum('transaction_type', TRANSACTION_TYPES);

export const usersTable = pgTable('users', {
	id: serial('id').primaryKey(),
	email: text('email').notNull().unique(),
	username: text('username').notNull().unique(),
	passwordHash: text('password_hash').notNull(),
	...timestamps
});

export const sessionsTable = pgTable('sessions', {
	id: text('id').primaryKey(),
	userId: integer('user_id')
		.notNull()
		.references(() => usersTable.id),
	expiresAt: timestamp('expires_at', dateParams).notNull(),
	...timestamps
});

export const ledgersTable = pgTable(
	'ledgers',
	{
		id: serial('id').primaryKey(),
		userId: integer('user_id')
			.notNull()
			.references(() => usersTable.id),
		title: text('title'),
		description: text('description'),
		year: integer('year').notNull(),
		month: integer('month').notNull(),
		type: ledgerTypeEnum('type').notNull(),
		status: ledgerStatusEnum('status').notNull().default('n/a'),
		totalCredits: numeric('total_credits', numericParams).notNull().default('0'),
		totalDebits: numeric('total_debits', numericParams).notNull().default('0'),
		...timestamps
	},
	(table) => [
		uniqueIndex('user_budget_month_year_unique_idx')
			.on(table.userId, table.year, table.month)
			.where(sql`${table.type} = 'budget'`)
	]
);

export const transactionsTable = pgTable('transactions', {
	id: serial('id').primaryKey(),
	ledgerId: integer('ledger_id')
		.notNull()
		.references(() => ledgersTable.id),
	description: text('description').notNull(),
	amount: numeric('amount', numericParams).notNull(),
	date: timestamp('date', dateParams).notNull(),
	category: transactionCategoryEnum('category').notNull(),
	type: transactionTypeEnum('type').notNull(),
	isEstimated: boolean('is_estimated').notNull().default(false),
	...timestamps
});

export type User = typeof usersTable.$inferSelect;
export type Session = typeof sessionsTable.$inferSelect;
export type Ledger = typeof ledgersTable.$inferSelect;
export type Transaction = typeof transactionsTable.$inferSelect;
