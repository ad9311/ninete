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
			return 'text-black';
		}

		switch (type) {
			case 'credit':
				return 'text-green-700';
			case 'debit':
				return 'text-red-700';
			case 'balance':
				return numericAmount > 0 ? 'text-green-700' : 'text-red-700';
			default:
				return 'text-black';
		}
	};
</script>

<span
	class="font-number inline-block rounded-none border border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-200 px-2 py-0.5 text-xs {getTextColor()}"
	>{formatedAmount}</span
>
