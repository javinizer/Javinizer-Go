<script lang="ts">
	import type { OrganizePreviewResponse } from '$lib/api/types';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { FolderOpen } from 'lucide-svelte';

	type OrganizeOperation = 'move' | 'copy' | 'hardlink' | 'softlink';

	interface Props {
		destinationPath: string;
		organizeOperation: OrganizeOperation;
		preview: OrganizePreviewResponse | null;
		showAllPreviewScreenshots: boolean;
		onOpenDestinationBrowser: () => void;
	}

	let {
		destinationPath = $bindable(''),
		organizeOperation = $bindable<OrganizeOperation>('move'),
		preview,
		showAllPreviewScreenshots = $bindable(false),
		onOpenDestinationBrowser
	}: Props = $props();
</script>

<Card class="p-4">
	<div class="space-y-3 min-w-0">
		<div class="flex items-center gap-2">
			<FolderOpen class="h-5 w-5 text-primary" />
			<h3 class="font-semibold">Output Destination</h3>
		</div>
		<div class="flex gap-2 min-w-0">
			<input
				type="text"
				bind:value={destinationPath}
				placeholder="Enter destination path (e.g., /path/to/output)"
				class="flex-1 min-w-0 px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all font-mono text-sm"
				title={destinationPath}
			/>
			<Button onclick={onOpenDestinationBrowser} variant="outline">
				{#snippet children()}
					<FolderOpen class="h-4 w-4 mr-2" />
					Browse
				{/snippet}
			</Button>
		</div>

		<div class="space-y-2">
			<label for="organizeOperation" class="text-sm font-medium">File operation</label>
			<select
				id="organizeOperation"
				bind:value={organizeOperation}
				class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all text-sm"
			>
				<option value="move">Move files</option>
				<option value="copy">Copy files</option>
				<option value="hardlink">Hard link files</option>
				<option value="softlink">Soft link files</option>
			</select>
			<p class="text-xs text-muted-foreground">
				{#if organizeOperation === 'hardlink'}
					Hard links require source and destination on the same filesystem.
				{:else if organizeOperation === 'softlink'}
					Soft links point to the original file path. Windows may require Developer Mode or elevated privileges.
				{:else if organizeOperation === 'copy'}
					Copy creates independent destination files and keeps originals unchanged.
				{:else}
					Move relocates source files into the organized destination.
				{/if}
			</p>
		</div>

		{#if preview}
			{@const pathParts = preview.full_path
				.replace(destinationPath + '/', '')
				.split('/')
				.filter((p) => p && !p.includes('.mp4'))}
			{@const fileIndent = pathParts.length * 4}
			<div class="mt-3 p-3 bg-accent/50 rounded border border-dashed overflow-hidden">
				<p class="text-xs font-medium mb-2 text-muted-foreground">Preview:</p>
				<div class="font-mono text-xs space-y-1 overflow-x-auto">
					<div class="text-muted-foreground break-all">📁 {destinationPath}/</div>
					{#each pathParts as part, index}
						<div class="text-muted-foreground break-all" style="margin-left: {(index + 1) * 4}px">📁 {part}/</div>
					{/each}
					{#if preview.video_files && preview.video_files.length > 0}
						{#each preview.video_files as videoFile}
							{@const fileName = videoFile.split(/[\\/]/).pop()}
							<div class="break-all" style="margin-left: {fileIndent + 4}px">🎬 {fileName}</div>
						{/each}
					{:else}
						<div class="break-all" style="margin-left: {fileIndent + 4}px">🎬 {preview.file_name}.mp4</div>
					{/if}
					{#if preview.nfo_path || (preview.nfo_paths && preview.nfo_paths.length > 0)}
						{#if preview.nfo_paths && preview.nfo_paths.length > 0}
							{#each preview.nfo_paths as nfoFile}
								{@const fileName = nfoFile.split(/[\\/]/).pop()}
								<div class="break-all" style="margin-left: {fileIndent + 4}px">📄 {fileName}</div>
							{/each}
						{:else if preview.nfo_path}
							<div class="break-all" style="margin-left: {fileIndent + 4}px">📄 {preview.nfo_path.split(/[\\/]/).pop()}</div>
						{/if}
					{/if}
					{#if preview.poster_path}
						<div class="break-all" style="margin-left: {fileIndent + 4}px">🖼️ {preview.poster_path.split(/[\\/]/).pop()}</div>
					{/if}
					{#if preview.fanart_path}
						<div class="break-all" style="margin-left: {fileIndent + 4}px">🖼️ {preview.fanart_path.split(/[\\/]/).pop()}</div>
					{/if}
					{#if preview.screenshots && preview.screenshots.length > 0}
						<div class="text-muted-foreground break-all" style="margin-left: {fileIndent + 4}px">📁 extrafanart/</div>
						{#each (showAllPreviewScreenshots ? preview.screenshots : preview.screenshots.slice(0, 3)) as screenshot}
							<div class="break-all" style="margin-left: {fileIndent + 8}px">🖼️ {screenshot}</div>
						{/each}
						{#if preview.screenshots.length > 3 && !showAllPreviewScreenshots}
							<button
								onclick={() => (showAllPreviewScreenshots = true)}
								class="text-muted-foreground hover:text-primary transition-colors cursor-pointer text-left"
								style="margin-left: {fileIndent + 8}px"
							>
								... and {preview.screenshots.length - 3} more
							</button>
						{/if}
						{#if showAllPreviewScreenshots && preview.screenshots.length > 3}
							<button
								onclick={() => (showAllPreviewScreenshots = false)}
								class="text-muted-foreground hover:text-primary transition-colors cursor-pointer text-left"
								style="margin-left: {fileIndent + 8}px"
							>
								Show less
							</button>
						{/if}
					{/if}
				</div>
			</div>
		{:else}
			<p class="text-xs text-muted-foreground">Files will be organized with metadata, artwork, and NFO files in this directory</p>
		{/if}
	</div>
</Card>
