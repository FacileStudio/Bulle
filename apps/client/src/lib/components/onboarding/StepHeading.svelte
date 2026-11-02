<script lang="ts">
	let { title, description = '' }: { title: string; description?: string } = $props();
</script>

<div class="flex max-w-fc-sm flex-col items-center gap-2 text-center">
	<h1
		data-step-heading
		data-reveal
		tabindex="-1"
		class="font-fc-title text-fc-2xl font-semibold tracking-tight text-fc-fg outline-none sm:text-fc-3xl"
	>
		{title}
	</h1>
	{#if description}
		<p data-reveal class="text-fc-md text-fc-fg-muted">{description}</p>
	{/if}
</div>
