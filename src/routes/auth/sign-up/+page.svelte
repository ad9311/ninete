<script lang="ts">
	import { enhance } from "$app/forms";
</script>

<form method="POST" use:enhance>
  <label for="username">Username</label>
  <input type="text" name="username" id="username" />

  <label for="email">Email</label>
  <input type="email" name="email" id="email" />

  <label for="password">Password</label>
  <input type="password" name="password" id="password" />

  <label for="password-confirmation">Password Confirmation</label>
  <input type="password" name="password-confirmation" id="password-confirmation" />

  <button type="submit">Sign up</button>
</form>
