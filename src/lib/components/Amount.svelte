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
			return 'var(--color-muted-foreground)';
		}

		switch (type) {
			case 'credit':
				return 'var(--color-positive)';
			case 'debit':
				return 'var(--color-negative)';
			case 'balance':
				return numericAmount > 0 ? 'var(--color-positive)' : 'var(--color-negative)';
			default:
				return 'var(--color-muted-foreground)';
		}
	};
</script>

<span style="color: {getColor()}">{formatedAmount}</span>
