<script lang="ts">
	import { DownloadSection } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { DESKTOP_CHECKSUMS_URL, DOWNLOADS, INSTALL_METHODS, RELEASES_URL } from '$lib/downloads';

	let { headingLevel = 2 }: { headingLevel?: 2 | 3 | 4 } = $props();
</script>

<div class="flex min-w-0 flex-col gap-4">
	<DownloadSection
		downloads={DOWNLOADS}
		commands={INSTALL_METHODS}
		releasesHref={RELEASES_URL}
		checksumsHref={DESKTOP_CHECKSUMS_URL}
		cliHint={t('downloads.terminalHint')}
		{headingLevel}
	/>
	<p class="text-fc-xs leading-relaxed text-fc-fg-muted">{t('downloads.notarized')}</p>
</div>
