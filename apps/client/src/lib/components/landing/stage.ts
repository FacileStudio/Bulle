import { cubicOut } from 'svelte/easing';

export type Script = (signal: AbortSignal) => Promise<void>;

/** Resolves after `ms`, or rejects as soon as `signal` aborts. */
export function sleep(ms: number, signal: AbortSignal): Promise<void> {
	return new Promise((resolve, reject) => {
		if (signal.aborted) return reject(signal.reason);
		const stop = () => {
			clearTimeout(id);
			reject(signal.reason);
		};
		const id = setTimeout(() => {
			signal.removeEventListener('abort', stop);
			resolve();
		}, ms);
		signal.addEventListener('abort', stop, { once: true });
	});
}

/** Reveals `text` one character at a time through `set`. */
export async function typeOut(text: string, set: (value: string) => void, interval: number, signal: AbortSignal) {
	for (let i = 1; i <= text.length; i++) {
		set(text.slice(0, i));
		await sleep(interval, signal);
	}
}

/** Runs `script` back to back until `signal` aborts; the abort itself is not an error. */
export async function loop(script: Script, signal: AbortSignal) {
	while (!signal.aborted) {
		try {
			await script(signal);
		} catch (error) {
			if (!signal.aborted) throw error;
		}
	}
}

/** True when the visitor asked the OS for reduced motion. */
export function prefersReducedMotion(): boolean {
	return typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Attachment that loops `script` while the node is on screen. Leaving the viewport aborts
 * the run and calls `onstop`, so the next visit starts from the opening frame.
 */
export function playInView(script: Script, onstop: () => void) {
	return (node: HTMLElement) => {
		if (prefersReducedMotion()) return;
		let controller: AbortController | null = null;
		const observer = new IntersectionObserver(
			([entry]) => {
				if (entry.isIntersecting && !controller) {
					controller = new AbortController();
					void loop(script, controller.signal);
				} else if (!entry.isIntersecting && controller) {
					controller.abort();
					controller = null;
					onstop();
				}
			},
			{ threshold: 0.35 }
		);
		observer.observe(node);
		return () => {
			observer.disconnect();
			controller?.abort();
		};
	};
}

/** Height-and-opacity transition: the room opens first, the content fades in as it lands. */
export function grow(node: Element, { delay = 0, duration = 450 }: { delay?: number; duration?: number } = {}) {
	const height = node.getBoundingClientRect().height;
	return {
		delay,
		duration,
		easing: cubicOut,
		css: (t: number) => `overflow: hidden; height: ${t * height}px; opacity: ${Math.max(0, (t - 0.4) / 0.6)};`
	};
}
