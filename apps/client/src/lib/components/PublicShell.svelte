<script lang="ts">
	import type { Snippet } from 'svelte';
	import { LandingFooter, LandingHeader, SignInWithFacile } from '@facile/muse';
	import { ssoLoginUrl } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { LOGO_ICON } from '$lib/icons';

	let { children }: { children: Snippet } = $props();

	const links = $derived([{ href: '/downloads', label: t('landing.downloads') }]);
	const columns = $derived([
		{
			title: t('landing.footer.product'),
			links: [
				{ href: '/downloads', label: t('landing.downloads') },
				{ href: 'https://github.com/FacileStudio/Bulle', label: 'GitHub', external: true }
			]
		}
	]);
</script>

<div class="flex min-h-dvh flex-col bg-fc-page text-fc-fg">
	<LandingHeader appName="Bulle" icon={LOGO_ICON} {links}>
		{#snippet actions()}
			<SignInWithFacile href={ssoLoginUrl} />
		{/snippet}
	</LandingHeader>

	<main class="flex-1">
		{@render children()}
	</main>

	<LandingFooter appName="Bulle" icon={LOGO_ICON} tagline={t('landing.tagline')} {columns} />
</div>
