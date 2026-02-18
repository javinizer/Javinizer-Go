<script lang="ts">
	import { cubicOut } from 'svelte/easing';
	import { fade, fly } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	interface Props {
		title: string;
		description?: string;
		children: Snippet;
	}

	let {
		title,
		description,
		children
	}: Props = $props();
</script>

<div
	class="settings-subsection mt-6 first:mt-0"
	in:fly|local={{ y: 6, duration: 220, easing: cubicOut }}
	out:fade|local={{ duration: 140 }}
>
	<div class="subsection-header mb-4">
		<h4 class="text-base font-semibold text-foreground">{title}</h4>
		{#if description}
			<p class="text-sm text-muted-foreground mt-1">{description}</p>
		{/if}
	</div>
	<div class="subsection-content space-y-0">
		{@render children()}
	</div>
</div>
