<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { cubicOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';
	import favicon from '$lib/assets/favicon.svg';
	import Navigation from '$lib/components/Navigation.svelte';
	import ToastContainer from '$lib/components/ui/ToastContainer.svelte';
	import { websocketStore } from '$lib/stores/websocket';
	import '../app.css';

	let { children } = $props();

	// Initialize WebSocket connection on mount
	onMount(() => {
		websocketStore.connect();
	});

	// Clean up WebSocket on destroy
	onDestroy(() => {
		websocketStore.disconnect();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="min-h-screen bg-background">
	<Navigation />
	<main class="route-container">
		{#key page.url.pathname}
			<div
				class="route-content"
				in:fly|local={{ y: 12, duration: 220, opacity: 0.15, easing: cubicOut }}
				out:fade|local={{ duration: 130 }}
			>
				{@render children?.()}
			</div>
		{/key}
	</main>
	<ToastContainer />
</div>

<style>
	.route-container {
		position: relative;
	}

	.route-content {
		will-change: auto;
	}
</style>
