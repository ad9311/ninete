import { ZodError } from 'zod';

export type Action = {
	label: string;
	href?: string;
	form?: string;
	submit?: boolean;
	className?: string;
	onClick?: () => void;
};

export const TRANSACTION_CATEGORIES = [
	'housing',
	'utilities',
	'groceries',
	'restaurants',
	'foodDelivery',
	'transportation',
	'healthcare&wellness',
	'personalCare',
	'shopping',
	'entertainment',
	'travel&vacations',
	'education',
	'children&dependents',
	'pets',
	'gifts&donations',
	'financialServices',
	'savings&investments',
	'workExpenses',
	'homeImprovement',
	'taxes',
	'miscellaneous',
	'income',
	'payable',
	'receivable'
] as const;
export type TRANSACTION_CATEGORY = (typeof TRANSACTION_CATEGORIES)[number];

export const LEDGER_TYPES = ['budget', 'payable', 'receivable', 'savings'] as const;
export type LEDGER_TYPE = (typeof LEDGER_TYPES)[number];

export const LEDGER_STATUS = ['n/a', 'pending', 'paid', 'overdue', 'cancelled'] as const;

export const TRANSACTION_TYPES = ['credit', 'debit'] as const;
export type TRANSACTION_TYPE = (typeof TRANSACTION_TYPES)[number];

export function formatFormErrors(error: Error | ZodError | undefined): string[] {
	if (!error) {
		return [''];
	}

	if (error instanceof ZodError) {
		return error.issues.map((issue) => {
			if (issue.path.length > 0) {
				return `${issue.path.join('.')}: ${issue.message}`;
			}
			return issue.message;
		});
	}

	if (error instanceof Error) {
		return [error.message];
	}

	return ['An unknown error occurred.'];
}

export function formatDateToMonthYear(
	date: Date,
	options: { includeDay: boolean } = { includeDay: false }
): string {
	const day = String(date.getDate()).padStart(2, '0');
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const year = date.getFullYear();

	return options.includeDay ? `${day}/${month}/${year}` : `${month}/${year}`;
}

export function formatMonthYear(month: number, year: number): string {
	const paddedMonth = String(month).padStart(2, '0');
	return `${paddedMonth}/${year}`;
}
