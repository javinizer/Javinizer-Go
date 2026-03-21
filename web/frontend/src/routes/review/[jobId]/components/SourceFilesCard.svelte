<script lang="ts">
	import type { FileResult } from '$lib/api/types';
	import Card from '$lib/components/ui/Card.svelte';
	import { truncatePath } from '../review-utils';

	interface Props {
		sourceResults: FileResult[];
		primaryFilePath: string;
		showFullSourcePath: boolean;
	}

	let { sourceResults, primaryFilePath, showFullSourcePath = $bindable(false) }: Props = $props();
</script>

<Card class="p-4">
	<div class="min-w-0">
		<div class="flex items-center justify-between mb-2">
			<p class="text-sm font-medium">
				{#if sourceResults.length > 1}
					Source Files ({sourceResults.length} parts)
				{:else}
					Source File
				{/if}
			</p>
			{#if primaryFilePath.length > 80}
				<button
					onclick={() => (showFullSourcePath = !showFullSourcePath)}
					class="text-xs text-primary hover:text-primary/80 transition-colors cursor-pointer"
				>
					{showFullSourcePath ? 'Hide' : 'Show full path'}
				</button>
			{/if}
		</div>
		{#if sourceResults.length > 1}
			<div class="space-y-2">
				{#each sourceResults as result, index}
					<div class="bg-accent rounded px-3 py-2 {showFullSourcePath ? 'overflow-x-auto' : ''}">
						<code class="text-xs block {showFullSourcePath ? 'whitespace-nowrap' : ''}" title={result.file_path}>
							<span class="text-muted-foreground mr-2">Part {index + 1}:</span>
							{showFullSourcePath ? result.file_path : truncatePath(result.file_path)}
						</code>
					</div>
				{/each}
			</div>
		{:else}
			<div class="bg-accent rounded px-3 py-2 {showFullSourcePath ? 'overflow-x-auto' : ''}">
				<code class="text-xs block {showFullSourcePath ? 'whitespace-nowrap' : ''}" title={primaryFilePath}>
					{showFullSourcePath ? primaryFilePath : truncatePath(primaryFilePath)}
				</code>
			</div>
		{/if}
	</div>
</Card>
