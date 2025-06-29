<script lang="ts">
	import { enhance } from '$app/forms';
</script>

<form method="post" class="form" use:enhance>
	<div class="form-group">
		<label for="title">Title</label>
		<input type="text" id="title" name="title" placeholder="Money I owe" />
	</div>
	<div class="form-group">
		<label for="description">Description</label>
		<input type="text" id="description" name="description" placeholder="300 dollars..." />
	</div>
	<div class="form-group">
		<label for="date">Date</label>
		<input type="date" id="Date" name="date" />
	</div>
	<div class="form-actions">
		<button type="submit" class="btn-primary">Create</button>
	</div>
</form>
