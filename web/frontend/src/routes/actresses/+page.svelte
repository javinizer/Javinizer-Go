<script lang="ts">
	import { onMount } from 'svelte';
	import { flip } from 'svelte/animate';
	import { quintOut } from 'svelte/easing';
	import { fade } from 'svelte/transition';
	import { Plus, RefreshCw, Search, Save, Trash2, Pencil, ImageOff, ChevronLeft, ChevronRight, ArrowUpDown } from 'lucide-svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { apiClient } from '$lib/api/client';
	import { toastStore } from '$lib/stores/toast';
	import type { Actress, ActressUpsertRequest } from '$lib/api/types';

	type ActressForm = {
		dmm_id: string;
		first_name: string;
		last_name: string;
		japanese_name: string;
		thumb_url: string;
		aliases: string;
	};

	const DEFAULT_LIMIT = 20;

	let actresses = $state<Actress[]>([]);
	let loading = $state(true);
	let saving = $state(false);
	let deletingId = $state<number | null>(null);
	let error = $state<string | null>(null);

	let queryInput = $state('');
	let activeQuery = $state('');
	let limit = $state(DEFAULT_LIMIT);
	let offset = $state(0);
	let total = $state(0);
	let viewMode = $state<'cards' | 'compact' | 'table'>('cards');
	let sortBy = $state<'name' | 'japanese_name' | 'id' | 'dmm_id' | 'updated_at' | 'created_at'>('name');
	let sortOrder = $state<'asc' | 'desc'>('asc');

	let editingId = $state<number | null>(null);
	let form = $state<ActressForm>(emptyForm());
	let formError = $state<string | null>(null);

	const currentPage = $derived(Math.floor(offset / limit) + 1);
	const totalPages = $derived(Math.max(1, Math.ceil(total / limit)));
	const canGoPrev = $derived(offset > 0);
	const canGoNext = $derived(offset + limit < total);

	function emptyForm(): ActressForm {
		return {
			dmm_id: '',
			first_name: '',
			last_name: '',
			japanese_name: '',
			thumb_url: '',
			aliases: ''
		};
	}

	function getDisplayName(actress: Actress): string {
		if (actress.last_name && actress.first_name) return `${actress.last_name} ${actress.first_name}`;
		if (actress.first_name) return actress.first_name;
		if (actress.japanese_name) return actress.japanese_name;
		return 'Unnamed';
	}

	function toPayload(): ActressUpsertRequest {
		const dmmID = Number.parseInt(form.dmm_id, 10);
		return {
			dmm_id: Number.isNaN(dmmID) ? undefined : dmmID,
			first_name: form.first_name.trim(),
			last_name: form.last_name.trim(),
			japanese_name: form.japanese_name.trim(),
			thumb_url: form.thumb_url.trim(),
			aliases: form.aliases.trim()
		};
	}

	function validateForm(): string | null {
		const payload = toPayload();
		if (payload.dmm_id !== undefined && payload.dmm_id < 0) {
			return 'DMM ID must be greater than or equal to 0';
		}
		if (!payload.first_name && !payload.japanese_name) {
			return 'Either First Name or Japanese Name is required';
		}
		return null;
	}

	function resetForm() {
		editingId = null;
		form = emptyForm();
		formError = null;
	}

	function startEdit(actress: Actress) {
		editingId = actress.id ?? null;
		form = {
			dmm_id: actress.dmm_id?.toString() ?? '',
			first_name: actress.first_name ?? '',
			last_name: actress.last_name ?? '',
			japanese_name: actress.japanese_name ?? '',
			thumb_url: actress.thumb_url ?? '',
			aliases: actress.aliases ?? ''
		};
		formError = null;
	}

	async function loadActresses() {
		loading = true;
		error = null;
		try {
			const response = await apiClient.listActresses({
				limit,
				offset,
				q: activeQuery || undefined,
				sort_by: sortBy,
				sort_order: sortOrder
			});
			actresses = response.actresses;
			total = response.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load actresses';
			actresses = [];
			total = 0;
		} finally {
			loading = false;
		}
	}

	async function saveActress() {
		formError = validateForm();
		if (formError) {
			return;
		}

		saving = true;
		try {
			const payload = toPayload();
			if (editingId !== null) {
				await apiClient.updateActress(editingId, payload);
				toastStore.success('Actress updated');
			} else {
				await apiClient.createActress(payload);
				toastStore.success('Actress created');
			}
			await loadActresses();
			resetForm();
		} catch (e) {
			formError = e instanceof Error ? e.message : 'Failed to save actress';
		} finally {
			saving = false;
		}
	}

	async function removeActress(actress: Actress) {
		if (!actress.id) return;
		const name = getDisplayName(actress);
		if (!confirm(`Delete actress "${name}"?`)) return;

		deletingId = actress.id;
		try {
			await apiClient.deleteActress(actress.id);
			toastStore.success('Actress deleted');
			if (actresses.length === 1 && offset > 0) {
				offset = Math.max(0, offset - limit);
			}
			await loadActresses();
			if (editingId === actress.id) {
				resetForm();
			}
		} catch (e) {
			toastStore.error(e instanceof Error ? e.message : 'Failed to delete actress');
		} finally {
			deletingId = null;
		}
	}

	function applySearch() {
		activeQuery = queryInput.trim();
		offset = 0;
		loadActresses();
	}

	function clearSearch() {
		queryInput = '';
		activeQuery = '';
		offset = 0;
		loadActresses();
	}

	function applySort() {
		offset = 0;
		loadActresses();
	}

	function toggleSortOrder() {
		sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		applySort();
	}

	function prevPage() {
		if (!canGoPrev) return;
		offset = Math.max(0, offset - limit);
		loadActresses();
	}

	function nextPage() {
		if (!canGoNext) return;
		offset += limit;
		loadActresses();
	}

	onMount(() => {
		loadActresses();
	});
</script>

<div class="container mx-auto px-4 py-8">
	<div class="max-w-7xl mx-auto space-y-6">
		<div class="flex flex-wrap items-center justify-between gap-3">
			<div>
				<h1 class="text-3xl font-bold">Actress Database</h1>
				<p class="text-muted-foreground mt-1">Create, update, and remove actress records stored in the database.</p>
			</div>
			<div class="flex items-center gap-2">
				<Button variant="outline" onclick={loadActresses}>
					<RefreshCw class="h-4 w-4" />
					Refresh
				</Button>
				<Button onclick={resetForm}>
					<Plus class="h-4 w-4" />
					New Actress
				</Button>
			</div>
		</div>

		<div class="grid grid-cols-1 xl:grid-cols-5 gap-6">
			<div class="xl:col-span-2">
				<Card class="p-5 space-y-4 sticky top-20">
					<h2 class="text-lg font-semibold">{editingId ? 'Edit Actress' : 'Create Actress'}</h2>

					{#if formError}
						<div class="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
							{formError}
						</div>
					{/if}

					<div class="space-y-3">
						<div>
							<label class="text-sm font-medium" for="dmm-id">DMM ID</label>
							<input
								id="dmm-id"
								type="number"
								min="0"
								bind:value={form.dmm_id}
								placeholder="e.g. 1092662"
								class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							/>
						</div>

						<div class="grid grid-cols-2 gap-3">
							<div>
								<label class="text-sm font-medium" for="first-name">First Name</label>
								<input
									id="first-name"
									type="text"
									bind:value={form.first_name}
									placeholder="Yui"
									class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
								/>
							</div>
							<div>
								<label class="text-sm font-medium" for="last-name">Last Name</label>
								<input
									id="last-name"
									type="text"
									bind:value={form.last_name}
									placeholder="Hatano"
									class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
								/>
							</div>
						</div>

						<div>
							<label class="text-sm font-medium" for="ja-name">Japanese Name</label>
							<input
								id="ja-name"
								type="text"
								bind:value={form.japanese_name}
								placeholder="波多野結衣"
								class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							/>
						</div>

						<div>
							<label class="text-sm font-medium" for="thumb-url">Thumbnail URL</label>
							<input
								id="thumb-url"
								type="url"
								bind:value={form.thumb_url}
								placeholder="https://example.com/actress.jpg"
								class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							/>
						</div>

						<div>
							<label class="text-sm font-medium" for="aliases">Aliases</label>
							<input
								id="aliases"
								type="text"
								bind:value={form.aliases}
								placeholder="Alias1|Alias2"
								class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
							/>
						</div>
					</div>

					<div class="flex items-center gap-2 pt-2">
						<Button onclick={saveActress} disabled={saving}>
							<Save class="h-4 w-4" />
							{saving ? 'Saving...' : editingId ? 'Update' : 'Create'}
						</Button>
						<Button variant="outline" onclick={resetForm} disabled={saving}>Clear</Button>
					</div>
				</Card>
			</div>

			<div class="xl:col-span-3 space-y-4">
				<Card class="p-4">
					<div class="flex flex-wrap items-center gap-2">
						<div class="flex-1 min-w-[220px]">
							<label class="sr-only" for="search">Search actresses</label>
							<div class="relative">
								<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
								<input
									id="search"
									type="text"
									bind:value={queryInput}
									onkeydown={(event) => {
										if (event.key === 'Enter') applySearch();
									}}
									placeholder="Search by English or Japanese name"
									class="w-full rounded-md border border-input bg-background pl-9 pr-3 py-2 text-sm"
								/>
							</div>
						</div>
						<Button onclick={applySearch}>Search</Button>
						<Button variant="outline" onclick={clearSearch}>Clear</Button>
					</div>
					<div class="mt-3 flex flex-wrap items-center justify-between gap-3">
						<div class="inline-flex rounded-md border border-input p-1">
							<Button
								size="sm"
								variant={viewMode === 'cards' ? 'default' : 'ghost'}
								onclick={() => {
									viewMode = 'cards';
								}}
							>
								Cards
							</Button>
							<Button
								size="sm"
								variant={viewMode === 'compact' ? 'default' : 'ghost'}
								onclick={() => {
									viewMode = 'compact';
								}}
							>
								Compact
							</Button>
							<Button
								size="sm"
								variant={viewMode === 'table' ? 'default' : 'ghost'}
								onclick={() => {
									viewMode = 'table';
								}}
							>
								Table
							</Button>
						</div>
						<div class="flex items-center gap-2">
							<select
								bind:value={sortBy}
								onchange={applySort}
								class="rounded-md border border-input bg-background px-3 py-2 text-sm"
								aria-label="Sort actresses by"
							>
								<option value="name">Sort: Name</option>
								<option value="japanese_name">Sort: Japanese Name</option>
								<option value="id">Sort: Database ID</option>
								<option value="dmm_id">Sort: DMM ID</option>
								<option value="updated_at">Sort: Updated Time</option>
								<option value="created_at">Sort: Created Time</option>
							</select>
							<Button variant="outline" size="sm" onclick={toggleSortOrder} title="Toggle sort direction">
								<ArrowUpDown class="h-4 w-4" />
								{sortOrder === 'asc' ? 'Asc' : 'Desc'}
							</Button>
						</div>
					</div>
					<div class="mt-3 text-sm text-muted-foreground">
						Showing {actresses.length} of {total} actress records
						{#if activeQuery}
							for "{activeQuery}"
						{/if}
					</div>
				</Card>

				{#if error}
					<Card class="p-4 border-destructive bg-destructive/10 text-destructive">{error}</Card>
				{/if}

				{#if loading}
					<Card class="p-8 text-center text-muted-foreground">Loading actresses...</Card>
				{:else if actresses.length === 0}
					<Card class="p-8 text-center">
						<p class="text-muted-foreground">No actresses found.</p>
					</Card>
				{:else}
					{#if viewMode === 'cards'}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
							{#each actresses as actress (actress.id)}
								<div animate:flip={{ duration: 220, easing: quintOut }} in:fade|local={{ duration: 140 }}>
									<Card class="p-3 h-full">
										<div class="flex items-start gap-3 h-full">
											{#if actress.thumb_url}
												<img
													src={actress.thumb_url}
													alt={getDisplayName(actress)}
													class="w-20 h-24 rounded object-cover border"
													onerror={(event) => {
														(event.currentTarget as HTMLImageElement).style.display = 'none';
													}}
												/>
											{:else}
												<div class="w-20 h-24 rounded border bg-muted flex items-center justify-center text-muted-foreground">
													<ImageOff class="h-4 w-4" />
												</div>
											{/if}

											<div class="flex-1 min-w-0">
												<div class="flex flex-wrap items-center gap-2">
													<h3 class="font-semibold truncate">{getDisplayName(actress)}</h3>
													{#if actress.id}
														<span class="text-xs rounded bg-muted px-2 py-0.5">#{actress.id}</span>
													{/if}
													{#if actress.dmm_id}
														<span class="text-xs rounded bg-muted px-2 py-0.5">DMM {actress.dmm_id}</span>
													{/if}
												</div>
												{#if actress.japanese_name}
													<p class="text-sm text-muted-foreground truncate">{actress.japanese_name}</p>
												{/if}
												{#if actress.aliases}
													<p class="text-xs text-muted-foreground line-clamp-2 mt-1">Aliases: {actress.aliases}</p>
												{/if}
												<div class="flex items-center gap-2 mt-3">
													<Button variant="outline" size="sm" onclick={() => startEdit(actress)}>
														<Pencil class="h-4 w-4" />
														Edit
													</Button>
													<Button
														variant="outline"
														size="sm"
														onclick={() => removeActress(actress)}
														disabled={deletingId === actress.id}
														class="text-destructive hover:bg-destructive/10"
													>
														<Trash2 class="h-4 w-4" />
														Delete
													</Button>
												</div>
											</div>
										</div>
									</Card>
								</div>
							{/each}
						</div>
					{:else if viewMode === 'compact'}
						<div class="space-y-2">
							{#each actresses as actress (actress.id)}
								<div animate:flip={{ duration: 220, easing: quintOut }} in:fade|local={{ duration: 140 }}>
									<Card class="p-3">
										<div class="flex items-center gap-3">
											<div class="flex-1 min-w-0">
												<div class="flex items-center gap-2 min-w-0">
													<p class="font-medium truncate">{getDisplayName(actress)}</p>
													{#if actress.id}
														<span class="text-xs rounded bg-muted px-2 py-0.5">#{actress.id}</span>
													{/if}
													{#if actress.dmm_id}
														<span class="text-xs rounded bg-muted px-2 py-0.5">DMM {actress.dmm_id}</span>
													{/if}
												</div>
												<p class="text-xs text-muted-foreground truncate">
													{actress.japanese_name || '-'}
												</p>
											</div>
											<div class="flex items-center gap-2">
												<Button variant="outline" size="sm" onclick={() => startEdit(actress)}>
													<Pencil class="h-4 w-4" />
												</Button>
												<Button
													variant="outline"
													size="sm"
													onclick={() => removeActress(actress)}
													disabled={deletingId === actress.id}
													class="text-destructive hover:bg-destructive/10"
												>
													<Trash2 class="h-4 w-4" />
												</Button>
											</div>
										</div>
									</Card>
								</div>
							{/each}
						</div>
					{:else}
						<Card class="overflow-hidden">
							<div class="overflow-x-auto">
								<table class="w-full text-sm">
									<thead class="bg-muted/50">
										<tr class="text-left border-b">
											<th class="px-3 py-2 font-medium">ID</th>
											<th class="px-3 py-2 font-medium">Name</th>
											<th class="px-3 py-2 font-medium">Japanese Name</th>
											<th class="px-3 py-2 font-medium">DMM ID</th>
											<th class="px-3 py-2 font-medium">Aliases</th>
											<th class="px-3 py-2 font-medium text-right">Actions</th>
										</tr>
									</thead>
									<tbody>
										{#each actresses as actress (actress.id)}
											<tr class="border-b last:border-b-0" in:fade|local={{ duration: 120 }}>
												<td class="px-3 py-2 text-muted-foreground">{actress.id ?? '-'}</td>
												<td class="px-3 py-2 font-medium max-w-44 truncate">{getDisplayName(actress)}</td>
												<td class="px-3 py-2 text-muted-foreground max-w-44 truncate">{actress.japanese_name || '-'}</td>
												<td class="px-3 py-2 text-muted-foreground">{actress.dmm_id ?? '-'}</td>
												<td class="px-3 py-2 text-muted-foreground max-w-52 truncate">{actress.aliases || '-'}</td>
												<td class="px-3 py-2">
													<div class="flex items-center justify-end gap-2">
														<Button variant="outline" size="sm" onclick={() => startEdit(actress)}>
															<Pencil class="h-4 w-4" />
														</Button>
														<Button
															variant="outline"
															size="sm"
															onclick={() => removeActress(actress)}
															disabled={deletingId === actress.id}
															class="text-destructive hover:bg-destructive/10"
														>
															<Trash2 class="h-4 w-4" />
														</Button>
													</div>
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</Card>
					{/if}
				{/if}

				<Card class="p-3">
					<div class="flex items-center justify-between text-sm">
						<div class="text-muted-foreground">
							Page {currentPage} of {totalPages}
						</div>
						<div class="flex items-center gap-2">
							<Button variant="outline" size="sm" onclick={prevPage} disabled={!canGoPrev}>
								<ChevronLeft class="h-4 w-4" />
								Prev
							</Button>
							<Button variant="outline" size="sm" onclick={nextPage} disabled={!canGoNext}>
								Next
								<ChevronRight class="h-4 w-4" />
							</Button>
						</div>
					</div>
				</Card>
			</div>
		</div>
	</div>
</div>
