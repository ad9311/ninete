import * as m from '$lib/paraglide/messages.js';
import { TRANSACTION_CATEGORIES } from '$lib/shared';

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

export function mapTransactionCategories(
	isPayableReceivable?: boolean
): { label: string; value: string }[] {
	if (isPayableReceivable) {
		return [
			{
				label: m['transactions.categories.payable'](),
				value: 'payable'
			},
			{
				label: m['transactions.categories.receivable'](),
				value: 'receivable'
			}
		];
	}

	return TRANSACTION_CATEGORIES.map((category) => {
		return {
			label: m[`transactions.categories.${category}`](),
			value: category
		};
	});
}

export function formatDateForInput(date: Date): string {
	return date.toISOString().split('T')[0];
}
