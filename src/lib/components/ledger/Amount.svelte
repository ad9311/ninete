<script lang="ts">
	export let type: 'credit' | 'debit' | 'balance';
	export let value: number | string;

	const numericAmount = Number.isNaN(+value) ? 0 : +value;

	const formatedAmount = new Intl.NumberFormat('es-CO', {
		style: 'currency',
		currency: 'COP'
	}).format(numericAmount);

	const getTextColor = () => {
		if (numericAmount === 0) {
			return 'text-muted bg-zinc-100 border-zinc-200';
		}

		switch (type) {
			case 'credit':
				return 'text-green-700 bg-green-100 border-green-200';
			case 'debit':
				return 'text-red-700 bg-red-100 border-red-200';
			case 'balance':
				return numericAmount > 0
					? 'text-green-700 bg-green-100 border-green-200'
					: 'text-red-700 bg-red-100 border-red-200';
			default:
				return 'text-muted bg-zinc-100 border-zinc-200';
		}
	};
</script>

<span class="{getTextColor()} rounded-xs border px-2 py-1">{formatedAmount}</span>
