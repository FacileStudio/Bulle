import { prefersReducedMotion } from '@facile/muse';
import gsap from 'gsap';

export function revealStep(root: HTMLElement | null, direction: 1 | -1): void {
	if (!root) return;
	gsap.killTweensOf(root);
	if (prefersReducedMotion()) {
		gsap.fromTo(root, { opacity: 0 }, { opacity: 1, duration: 0.15, clearProps: 'all' });
		return;
	}
	const parts = root.querySelectorAll('[data-reveal]');
	gsap.fromTo(
		parts,
		{ opacity: 0, x: direction * 28, y: 12 },
		{
			opacity: 1,
			x: 0,
			y: 0,
			duration: 0.55,
			stagger: 0.07,
			ease: 'power3.out',
			clearProps: 'all'
		}
	);
}

export function popAvatar(el: HTMLElement | null, strong: boolean): void {
	if (!el || prefersReducedMotion()) return;
	gsap.killTweensOf(el);
	gsap.fromTo(
		el,
		{ scale: strong ? 0.4 : 0.92, opacity: strong ? 0 : 1 },
		{ scale: 1, opacity: 1, duration: strong ? 0.9 : 0.5, ease: 'back.out(1.6)', clearProps: 'all' }
	);
}
