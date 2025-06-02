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

export const LEDGER_TYPES = ['budget', 'payable/receivable'] as const;

export const LEDGER_STATUS = ['n/a', 'pending', 'paid', 'overdue', 'cancelled'] as const;

export const TRANSACTION_TYPES = ['credit', 'debit'] as const;

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
