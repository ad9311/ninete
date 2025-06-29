import * as m from '$lib/paraglide/messages.js';
import { TRANSACTION_CATEGORIES, type TRANSACTION_CATEGORY } from '$lib/shared';

export type BreadcrumbItem = {
	label: string;
	href?: string;
};

/* eslint-disable @typescript-eslint/no-explicit-any */
export type NavLink = {
	label: string;
	path: string;
	active?: boolean;
	icon?: any;
};
/* eslint-enable @typescript-eslint/no-explicit-any */

export function mapTransactionCategories(isLoan?: boolean): { label: string; value: string }[] {
	const excludeCategories = ['payable', 'receivable'];
	const categories = isLoan
		? ['payment', 'loan']
		: TRANSACTION_CATEGORIES.filter((c) => !excludeCategories.includes(c));

	return categories.map((category) => {
		return {
			label: m[`transactions.categories.${category as TRANSACTION_CATEGORY}`](),
			value: category
		};
	});
}

export function formatDateForInput(date: Date): string {
	return date.toISOString().split('T')[0];
}
