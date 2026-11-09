<script lang="ts">
	import { Badge, Button, Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	const cards = [
		{
			kind: 'mail',
			icon: icons.mail,
			titleKey: 'landing.ideas.card1.title' as const,
			bodyKey: 'landing.ideas.card1.body' as const,
			app: 'Courrier'
		},
		{
			kind: 'task',
			icon: icons.check,
			titleKey: 'landing.ideas.card2.title' as const,
			bodyKey: 'landing.ideas.card2.body' as const,
			app: 'Boite VM'
		},
		{
			kind: 'event',
			icon: icons.calendar,
			titleKey: 'landing.ideas.card3.title' as const,
			bodyKey: 'landing.ideas.card3.body' as const,
			app: 'Agenda'
		}
	];
</script>

<section class="relative mx-auto max-w-5xl px-6 py-16 sm:py-20" aria-labelledby="ideas-heading">
	<div class="mb-10 text-center">
		<div class="mb-3 inline-flex items-center gap-2 rounded-full bg-fc-surface px-3 py-1 text-fc-xs font-medium text-fc-accent">
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

	<div class="flex flex-col gap-4 rounded-2xl bg-fc-component p-5 sm:p-7">
		<div class="flex flex-wrap items-center justify-between gap-3 border-b border-fc-line pb-4">
			<div class="flex items-center gap-2.5">
				<div class="flex size-8 items-center justify-center rounded-full bg-fc-accent/15 text-fc-accent">
					<Icon icon="solar:lightbulb-bold-duotone" size={18} />
				</div>
				<div class="flex items-center gap-2">
					<span class="font-fc-title text-fc-sm font-semibold text-fc-fg">{t('landing.ideas.morning')}</span>
					<span class="text-fc-xs text-fc-fg-muted">• 8:00 AM</span>
				</div>
			</div>
			<Badge tone="accent">3 fresh ideas</Badge>
		</div>

		<div class="grid gap-3.5 md:grid-cols-3">
			{#each cards as card (card.kind)}
				<div class="flex flex-col justify-between gap-4 rounded-xl bg-fc-surface p-4 transition-colors hover:bg-fc-field">
					<div class="flex flex-col gap-2.5">
						<div class="flex items-center justify-between gap-2">
							<span class="inline-flex size-7 items-center justify-center rounded-full bg-fc-page text-fc-accent">
								<Icon icon={card.icon} size={15} />
							</span>
							<span class="text-fc-xs font-medium text-fc-fg-muted">{card.app}</span>
						</div>
						<h3 class="font-fc-title text-fc-sm font-semibold leading-snug text-fc-fg">
							{t(card.titleKey)}
						</h3>
						<p class="text-fc-xs leading-relaxed text-fc-fg-muted">
							{t(card.bodyKey)}
						</p>
					</div>

					<div class="pt-1">
						<Button
							href="/login"
							variant="outline"
							size="sm"
							iconRight={icons.arrow}
							class="w-full justify-between rounded-fc-md text-fc-xs"
						>
							{t('landing.ideas.start')}
						</Button>
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>
