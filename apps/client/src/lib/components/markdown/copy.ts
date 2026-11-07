import { t } from '$lib/i18n/index.svelte';
import { toast } from '@facile/muse';

/** Copies text to the clipboard and says whether it worked. */
export async function copyText(text: string): Promise<void> {
	try {
		await navigator.clipboard.writeText(text);
		toast.success(t('markdown.copied'));
	} catch {
		toast.danger(t('markdown.copyFailed'));
	}
}
