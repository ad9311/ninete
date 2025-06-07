import * as m from '$lib/paraglide/messages.js';
import { TRANSACTION_CATEGORIES } from '$lib/shared';

export type BreadcrumbItem = {
	label: string;
	href?: string;
};

/* eslint-disable @typescript-eslint/no-explicit-any */
export type NavLink = {
	lable: string;
	path: string;
	active?: boolean;
	icon?: any;
};
/* eslint-enable @typescript-eslint/no-explicit-any */

export function mapTransactionCategories(): { label: string; value: string }[] {
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
