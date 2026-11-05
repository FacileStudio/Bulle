<script lang="ts">
	import { goto } from '$app/navigation';
	import { Alert, Badge, Button, Icon, Section, Spinner } from '@facile/muse';
	import { SUITE_APP_ICONS, SUITE_FALLBACK_ICON } from '$lib/suite-icons';
	import { signOut } from '$lib/session.svelte';
	import { badgeFor, getSuiteStatus, type SuiteApp } from '$lib/suite';

	let apps = $state<SuiteApp[]>([]);
	let loading = $state(true);
	let failed = $state<string | null>(null);

	$effect(() => {
		getSuiteStatus()
			.then((result) => (apps = result))
			.catch((err) => (failed = err instanceof Error ? err.message : 'Could not load connected apps'))
			.finally(() => (loading = false));
	});

	async function reconnect() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

{#if loading || failed || apps.length > 0}
	<Section
		card
		title="Connected apps"
		description="Your assistant reads your own data in these Facile apps. It never sees anyone else's."
	>
		<div class="flex flex-col gap-3">
			{#if loading}
				<Spinner size="sm" />
			{:else if failed}
				<Alert tone="danger">{failed}</Alert>
			{:else}
				{#each apps as item (item.app)}
					{@const badge = badgeFor(item.state)}
					<div class="flex flex-wrap items-center justify-between gap-3 py-1.5">
						<div class="flex min-w-0 items-center gap-3">
							<Icon
								icon={SUITE_APP_ICONS[item.app] ?? SUITE_FALLBACK_ICON}
								size={22}
								class="text-fc-fg"
							/>
							<span class="truncate text-fc-sm font-medium text-fc-fg">{item.name}</span>
							<Badge tone={badge.tone}>{badge.label}</Badge>
						</div>
						{#if item.state === 'no_account'}
							<a
								href={item.url}
								target="_blank"
								rel="noopener noreferrer"
								class="inline-flex min-h-11 items-center text-fc-sm font-medium text-fc-accent underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
							>
								Open {item.name} once to create your account
							</a>
						{:else if item.state === 'reconnect'}
							<Button variant="outline" onclick={reconnect}>Sign in again</Button>
						{/if}
					</div>
				{/each}
			{/if}
		</div>
	</Section>
{/if}
