<script lang="ts">
	import './layout.css';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import { checkSession, loading } from '$lib/stores/auth';

	let { children } = $props();

	const messages = [
		'Counting bytes… just kidding!',
		'Warming up the Yippee! engine...',
		'Trying our best to find your files...',
		'Need. More. Coffee. ASAP.',
		'Fact: 40% of all data is just cats wearing hats.',
		'"To err is human but to really foul things up requires a computer." - Paul R. Ehrlich',
		'"Keep your friends close, and your backups closer." - Anonymous',
	];

	let currentMessage = $state(messages[0]);

	onMount(() => {
		checkSession();
		const interval = setInterval(() => {
			const randomIndex = Math.floor(Math.random() * messages.length);
			currentMessage = messages[randomIndex];
		}, 3500);

		return () => clearInterval(interval);
	});
</script>

{#if $loading}
	<div
		class="flex h-screen w-full flex-col items-center justify-center gap-6 bg-base-100"
		in:fade={{ duration: 200 }}
		out:fade={{ duration: 200 }}
	>
		<span class="loading loading-lg scale-150 loading-bars text-primary"></span>

		<div class="flex h-6 items-center justify-center">
			{#key currentMessage}
				<p
					in:fade={{ duration: 300, delay: 100 }}
					class="text-center font-mono text-sm tracking-wide text-base-content/50 max-w-xs"
				>
					{currentMessage}
				</p>
			{/key}
		</div>
	</div>
{:else}
	{@render children()}
{/if}
