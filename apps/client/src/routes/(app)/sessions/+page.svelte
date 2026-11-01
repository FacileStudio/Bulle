<script lang="ts">
	import { EmptyState, PageHeader, Stack, icons } from '@facile/muse';
</script>

<svelte:head>
	<title>Sessions - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Sessions" description="Your conversations with the agent." />
	<EmptyState
		icon={icons.history}
		title="No sessions yet"
		description="Chat and the agent's computer arrive in the next release."
	/>
</Stack>
