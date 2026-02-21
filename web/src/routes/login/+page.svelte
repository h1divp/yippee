<script lang="ts">
	import Logo from '$lib/assets/yippee_logo.png';
	import { login } from '$lib/stores/auth';

	// In the future, this comes from a config file or something, but for now it's hardcoded
	// Maybe an env variable? Might be simplest.
	let instanceName = 'My Yippee! Instance';

	let username = '';
	let password = '';
	let isSubmitting = false;

	async function onSubmit() {
		// Validation
		if (!username.trim() || !password.trim()) {
			alert('Please enter a username and password');
			return;
		}

		isSubmitting = true;
		console.log('Logging in with', { username, password });

		// Timeout for a few secs
		setTimeout(async () => {
			await login(username, password);

			isSubmitting = false;
			alert('Logged in!');
		}, 2000);
	}
</script>

<div class="relative flex min-h-screen flex-col items-center justify-center p-4">
	<div class="mb-12 flex w-full max-w-xs flex-col items-center">
		<img src={Logo} alt="yippee! Logo" class="w-96" />
		<div class="mb-4 text-center">
			<h1 class="font-mono text-2xl font-bold tracking-tight">
				{instanceName}
			</h1>
		</div>

		<form class="w-full">
			<fieldset class="fieldset flex w-full flex-col gap-2">
				<input
					bind:value={username}
					type="text"
					class="input-bordered input w-full"
					placeholder="Username"
				/>
				<input
					bind:value={password}
					class="input-bordered input w-full"
					placeholder="Password"
					type="password"
				/>
				<button
					on:click={onSubmit}
					type="submit"
					class={`btn btn-primary ${isSubmitting && 'btn-disabled'}`}
					disabled={isSubmitting}
				>
					{#if isSubmitting}
						<span class="loading loading-spinner"></span>
					{:else}
						Log In
					{/if}
				</button>
			</fieldset>
		</form>
	</div>

	<div class="absolute bottom-4 text-center">
		<p class="font-mono text-xs leading-relaxed text-base-content/40">
			<a
				href="https://github.com/h1divp/yippee"
				target="_blank"
				rel="noreferrer"
				class="font-bold text-primary">yippee!</a
			>
			— The self-hosted file system.
		</p>
	</div>
</div>
