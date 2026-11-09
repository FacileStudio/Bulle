<script lang="ts">
	import { Button, Icon, icons } from '@facile/muse';
	import { SUITE_APP_ICONS } from '$lib/suite-icons';
	import { t } from '$lib/i18n/index.svelte';

	const cards = [
		{
			kind: 'mail',
			icon: SUITE_APP_ICONS.courrier,
			titleKey: 'landing.ideas.card1.title' as const,
			bodyKey: 'landing.ideas.card1.body' as const,
			app: 'Courrier'
		},
		{
			kind: 'task',
			icon: SUITE_APP_ICONS.boite,
			titleKey: 'landing.ideas.card2.title' as const,
			bodyKey: 'landing.ideas.card2.body' as const,
			app: 'Boite VM'
		},
		{
			kind: 'event',
			icon: SUITE_APP_ICONS.agenda,
			titleKey: 'landing.ideas.card3.title' as const,
			bodyKey: 'landing.ideas.card3.body' as const,
			app: 'Agenda'
		}
	];
</script>

<section class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20" aria-labelledby="ideas-heading">
	<div class="mb-10 text-center">
		<div class="mb-3 inline-flex items-center gap-1.5 rounded-full bg-fc-surface px-3 py-1 text-fc-xs font-medium text-fc-accent">
			<Icon icon="solar:lightbulb-bold-duotone" size={16} />
			<span>{t('landing.ideas.badge')}</span>
		</div>
		<h2 id="ideas-heading" class="font-fc-title text-2xl font-semibold tracking-tight text-fc-fg sm:text-3xl">
			{t('landing.ideas.title')}
		</h2>
		<p class="mx-auto mt-3 max-w-xl text-fc-sm text-fc-fg-muted">
			{t('landing.ideas.body')}
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-3">
		{#each cards as card (card.kind)}
			<article class="flex flex-col justify-between rounded-fc-lg bg-fc-component p-6 transition-colors hover:bg-fc-component/80">
				<div class="flex flex-col gap-3">
					<div class="flex items-center justify-between gap-3">
						<div class="flex items-center gap-2.5">
							<span class="flex size-9 shrink-0 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-accent">
								<Icon icon={card.icon} size={18} />
							</span>
							<span class="text-fc-xs font-semibold text-fc-fg">{card.app}</span>
						</div>
						<span class="rounded-full bg-fc-surface px-2.5 py-0.5 text-fc-xs font-medium text-fc-fg-muted">8:00 AM</span>
					</div>

					<h3 class="font-fc-title text-fc-base font-semibold leading-snug text-fc-fg">
						{t(card.titleKey)}
					</h3>
					<p class="text-fc-sm leading-relaxed text-fc-fg-muted">
						{t(card.bodyKey)}
					</p>
				</div>

				<div class="mt-6 pt-2">
					<Button
						href="/login"
						variant="ghost"
						size="sm"
						iconRight={icons.arrow}
						class="-ml-2 text-fc-accent hover:text-fc-fg"
					>
						{t('landing.ideas.start')}
					</Button>
				</div>
			</article>
		{/each}
	</div>
</section>
