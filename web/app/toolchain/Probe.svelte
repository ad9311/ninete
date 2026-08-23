<script lang="ts">
  // The smallest component that exercises every part of the pipeline a real one
  // will use: a typed prop, reactive state, an event handler and a derived
  // value. It is not part of the app and nothing imports it outside its test.
  let { label }: { label: string } = $props();

  let count = $state(0);
  const heading = $derived(`${label}: ${count}`);
</script>

<button type="button" onclick={() => count++}>{heading}</button>
