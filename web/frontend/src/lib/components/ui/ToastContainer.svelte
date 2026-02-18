<script lang="ts">
	import { flip } from 'svelte/animate';
	import { cubicOut } from 'svelte/easing';
	import { toastStore } from '$lib/stores/toast';
	import Toast from './Toast.svelte';

	const toasts = $derived($toastStore);
</script>

<div class="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
	<div class="flex flex-col gap-2 pointer-events-auto">
		{#each toasts as toast (toast.id)}
			<div animate:flip={{ duration: 220, easing: cubicOut }}>
				<Toast
					id={toast.id}
					type={toast.type}
					message={toast.message}
					duration={toast.duration}
					onDismiss={toastStore.dismiss}
				/>
			</div>
		{/each}
	</div>
</div>
