import type { Snippet } from 'svelte';

export type BreadcrumbItem = {
	label: string;
	href?: string;
};

export type NavLink = {
	lable: string;
	path: string;
	icon?: Snippet;
};
