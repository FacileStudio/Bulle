import { toast } from '@facile/muse';

/** Copies text to the clipboard and says whether it worked. */
export async function copyText(text: string): Promise<void> {
	try {
		await navigator.clipboard.writeText(text);
		toast.success('Copied to the clipboard');
	} catch {
		toast.danger('Could not copy');
	}
}
