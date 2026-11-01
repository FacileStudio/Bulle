<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button, PageHeader, ProfileCard, Stack, icons } from '@facile/muse';
	import { session, signOut } from '$lib/session.svelte';

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

<svelte:head>
	<title>Account - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Account" description="Your identity comes from Facile." />
	{#if session.user}
		<ProfileCard name={session.user.name.trim() || session.user.email} email={session.user.email}>
			{#snippet actions()}
				<Button variant="outline" icon={icons.logout} onclick={leave}>Sign out</Button>
			{/snippet}
		</ProfileCard>
	{/if}
</Stack>
