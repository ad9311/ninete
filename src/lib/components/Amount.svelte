<!-- <script lang="ts">
	export let type: 'credit' | 'debit' | 'balance';
	export let value: number | string;
	export let className: string = '';

	const numericAmount = Number.isNaN(+value) ? 0 : +value;

	const formatedAmount = new Intl.NumberFormat('es-CO', {
		style: 'currency',
		currency: 'COP'
	}).format(numericAmount);

	const getColor = () => {
		if (numericAmount === 0) {
			return 'secondary';
		}

		switch (type) {
			case 'credit':
				return 'positive';
			case 'debit':
				return 'destructive';
			case 'balance':
				return numericAmount > 0 ? 'positive' : 'destructive';
			default:
				return 'secondary';
		}
	};
</script> -->
