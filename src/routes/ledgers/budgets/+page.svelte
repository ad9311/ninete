<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';

	const breadcrumbItems: BreadcrumbItem[] = [{ label: 'Home', href: '/' }, { label: 'Budgets' }];
</script>

<Breadcrumb items={breadcrumbItems} />
