<script lang="ts">
	import { onMount } from 'svelte';
	import { quintOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';
	import { browser } from '$app/environment';
	import { apiClient } from '$lib/api/client';
	import { websocketStore } from '$lib/stores/websocket';
	import type { HealthResponse } from '$lib/api/types';

	let health: HealthResponse | null = $state(null);
	let error: string | null = $state(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			health = await apiClient.health();
			websocketStore.connect();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to connect to API';
		} finally {
			loading = false;
		}
	});

	const wsState = $derived($websocketStore);
</script>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-4xl mx-auto">
		<h1 class="text-4xl font-bold mb-8 text-center">Javinizer Web UI</h1>

		<div class="grid gap-6">
			<!-- API Status Card -->
			<div class="border rounded-lg p-6 shadow-sm" in:fly|local={{ y: 10, duration: 220, easing: quintOut }}>
				<h2 class="text-2xl font-semibold mb-4">API Status</h2>
				{#if loading}
					<p class="text-muted-foreground">Connecting to API...</p>
				{:else if error}
					<div class="bg-destructive/10 border border-destructive text-destructive px-4 py-3 rounded">
						<p class="font-semibold">Error</p>
						<p class="text-sm">{error}</p>
					</div>
				{:else if health}
					<div class="space-y-2">
						<div class="flex items-center gap-2">
							<div class="w-2 h-2 bg-green-500 rounded-full"></div>
							<span class="font-medium">Connected</span>
						</div>
						<div>
							<span class="text-sm text-muted-foreground">Available scrapers:</span>
							<div class="flex gap-2 mt-1">
								{#each health.scrapers as scraper}
									<span class="px-2 py-1 bg-primary/10 text-primary rounded text-sm" in:fade|local={{ duration: 140 }}>
										{scraper}
									</span>
								{/each}
							</div>
						</div>
					</div>
				{/if}
			</div>

			<!-- WebSocket Status Card -->
			<div class="border rounded-lg p-6 shadow-sm" in:fly|local={{ y: 10, duration: 250, easing: quintOut }}>
				<h2 class="text-2xl font-semibold mb-4">WebSocket Status</h2>
				<div class="flex items-center gap-2">
					<div class="w-2 h-2 {wsState.connected ? 'bg-green-500' : 'bg-red-500'} rounded-full"></div>
					<span class="font-medium">{wsState.connected ? 'Connected' : 'Disconnected'}</span>
				</div>
				{#if wsState.error}
					<p class="text-sm text-destructive mt-2">{wsState.error}</p>
				{/if}
				<p class="text-sm text-muted-foreground mt-2">
					Messages received: {wsState.messages.length}
				</p>
			</div>

			<!-- Navigation -->
			<div class="border rounded-lg p-6 shadow-sm" in:fly|local={{ y: 10, duration: 280, easing: quintOut }}>
				<h2 class="text-2xl font-semibold mb-4">Quick Start</h2>
				<p class="text-muted-foreground mb-4">
					Welcome to Javinizer! This web interface allows you to scrape JAV metadata and organize
					your video library.
				</p>
				<div class="grid gap-3">
					<a
						href="/browse"
						class="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors text-center"
					>
						Browse & Scrape Files
					</a>
					<a
						href="/settings"
						class="px-4 py-2 border rounded hover:bg-accent transition-colors text-center"
					>
						Settings
					</a>
					<a
						href={browser ? `${location.protocol}//${location.hostname}:8080/docs` : '/docs'}
						target="_blank"
						rel="noopener noreferrer"
						class="px-4 py-2 border rounded hover:bg-accent transition-colors text-center"
					>
						API Documentation
					</a>
				</div>
			</div>
		</div>
	</div>
</div>
