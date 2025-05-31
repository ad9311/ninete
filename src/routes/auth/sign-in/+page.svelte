<script lang="ts">
	import { enhance } from "$app/forms";
</script>

<form method="POST" use:enhance>
  <label for="email">Email</label>
  <input type="email" name="email" id="email" />

  <label for="password">Password</label>
  <input type="password" name="password" id="password" />

  <button type="submit">Sign in</button>
</form>
