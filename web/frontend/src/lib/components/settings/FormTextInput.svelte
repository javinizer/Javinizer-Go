<script lang="ts">
	interface Props {
		label: string;
		description?: string;
		value: string;
		placeholder?: string;
		onchange: (value: string) => void;
		disabled?: boolean;
		pattern?: string;
		id?: string;
	}

	let {
		label,
		description,
		value = $bindable(''),
		placeholder,
		onchange,
		disabled = false,
		pattern,
		id = `text-${crypto.randomUUID()}`
	}: Props = $props();

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		onchange(target.value);
	}
</script>

<div class="form-row py-4 border-b border-border last:border-0">
	<div class="form-label flex-1">
		<label for={id} class="text-sm font-medium text-foreground">
			{label}
		</label>
		{#if description}
			<p class="text-sm text-muted-foreground mt-1" id="{id}-desc">{description}</p>
		{/if}
	</div>
	<div class="form-control flex-1">
		<input
			type="text"
			{id}
			bind:value
			oninput={handleInput}
			{placeholder}
			{disabled}
			{pattern}
			aria-describedby={description ? `${id}-desc` : undefined}
			class="w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-primary focus:border-primary transition-all bg-background text-sm disabled:opacity-50"
		/>
	</div>
</div>

<style>
	.form-row {
		display: flex;
		align-items: start;
		gap: 1rem;
	}

	@media (max-width: 768px) {
		.form-row {
			flex-direction: column;
		}
	}
</style>
