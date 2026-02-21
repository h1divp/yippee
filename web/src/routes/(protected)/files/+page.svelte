<script>
	import { logout } from "$lib/stores/auth";

</script>
<h1 class="mb-4 ml-4 mt-4">This is your root (`/`)</h1>

<button on:click={logout} class="btn btn-error ml-4">Log Out</button>