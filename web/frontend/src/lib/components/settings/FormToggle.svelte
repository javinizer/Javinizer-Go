<script lang="ts">
	interface Props {
		label: string;
		description?: string;
		checked: boolean;
		onchange: (value: boolean) => void;
		disabled?: boolean;
		id?: string;
	}

	let {
		label,
		description,
		checked = $bindable(false),
		onchange,
		disabled = false,
		id = `toggle-${crypto.randomUUID()}`
	}: Props = $props();

	function handleChange(event: Event) {
		const target = event.target as HTMLInputElement;
		onchange(target.checked);
	}
</script>

<div class="form-row py-4 border-b border-border last:border-0">
	<div class="form-label flex-1">
		<label for={id} class="text-sm font-medium text-foreground cursor-pointer">
			{label}
		</label>
		{#if description}
			<p class="text-sm text-muted-foreground mt-1" id="{id}-desc">{description}</p>
		{/if}
	</div>
	<div class="form-control flex items-center ml-4">
		<input
			type="checkbox"
			{id}
			bind:checked
			onchange={handleChange}
			{disabled}
			aria-describedby={description ? `${id}-desc` : undefined}
			class="h-4 w-4 rounded border-gray-300 text-primary focus:ring-2 focus:ring-primary disabled:opacity-50 cursor-pointer"
		/>
	</div>
</div>

<style>
	.form-row {
		display: flex;
		align-items: start;
		gap: 1rem;
	}
</style>
