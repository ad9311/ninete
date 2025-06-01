<script lang="ts">
	export let type: 'credit' | 'debit' | 'balance';
	export let value: number | string;

	const numericAmount = Number.isNaN(+value) ? 0 : +value;

	const formatedAmount = new Intl.NumberFormat('es-CO', {
		style: 'currency',
		currency: 'COP'
	}).format(numericAmount);

	const getColor = () => {
		if (numericAmount === 0) {
			return 'text-muted';
		}

		switch (type) {
			case 'credit':
				return 'text-green-600';
			case 'debit':
				return 'text-red-600';
			case 'balance':
				return numericAmount > 0 ? 'text-green-600' : 'text-red-600';
			default:
				return 'text-muted';
		}
	};
</script>

<span class={getColor()}>{formatedAmount}</span>
