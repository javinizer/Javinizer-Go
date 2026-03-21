<script lang="ts">
	import { ChevronLeft, ChevronRight, CircleAlert, Trash2 } from 'lucide-svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';

	interface Props {
		currentMovieIndex: number;
		movieResultsLength: number;
		currentMovieId: string;
		hasChanges: boolean;
		onExclude: () => void;
	}

	let {
		currentMovieIndex = $bindable(0),
		movieResultsLength,
		currentMovieId,
		hasChanges,
		onExclude
	}: Props = $props();
</script>

<Card class="p-4">
	<div class="flex items-center justify-between">
		<Button
			variant="outline"
			onclick={() => (currentMovieIndex = Math.max(0, currentMovieIndex - 1))}
			disabled={currentMovieIndex === 0}
		>
			{#snippet children()}
				<ChevronLeft class="h-4 w-4 mr-2" />
				Previous
			{/snippet}
		</Button>

		<div class="text-center flex-1 mx-4">
			<p class="font-semibold">Movie {currentMovieIndex + 1} of {movieResultsLength}</p>
			<p class="text-sm text-muted-foreground">{currentMovieId}</p>
			{#if hasChanges}
				<span class="text-xs text-orange-600 flex items-center gap-1 justify-center mt-1">
					<CircleAlert class="h-3 w-3" />
					Modified
				</span>
			{/if}
		</div>

		<div class="flex gap-2">
			<Button
				variant="outline"
				onclick={onExclude}
				class="text-destructive hover:bg-destructive hover:text-destructive-foreground"
			>
				{#snippet children()}
					<Trash2 class="h-4 w-4 mr-2" />
					Remove
				{/snippet}
			</Button>

			<Button
				variant="outline"
				onclick={() => (currentMovieIndex = Math.min(movieResultsLength - 1, currentMovieIndex + 1))}
				disabled={currentMovieIndex === movieResultsLength - 1}
			>
				{#snippet children()}
					Next
					<ChevronRight class="h-4 w-4 ml-2" />
				{/snippet}
			</Button>
		</div>
	</div>
</Card>
