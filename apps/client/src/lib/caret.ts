export interface CaretCoordinates {
	x: number;
	y: number;
}

const PROPERTIES = [
	'borderTopWidth',
	'borderRightWidth',
	'borderBottomWidth',
	'borderLeftWidth',
	'paddingTop',
	'paddingRight',
	'paddingBottom',
	'paddingLeft',
	'fontFamily',
	'fontSize',
	'fontWeight',
	'fontStyle',
	'letterSpacing',
	'lineHeight',
	'textTransform',
	'wordSpacing',
	'textIndent',
	'tabSize'
] as const;

export function getCaretCoordinates(
	element?: HTMLTextAreaElement | null,
	position?: number
): CaretCoordinates {
	if (!element || typeof window === 'undefined' || typeof document === 'undefined') {
		return { x: 0, y: 0 };
	}

	const mirror = document.createElement('div');
	const style = window.getComputedStyle(element);

	const paddingLeft = parseFloat(style.paddingLeft || '0');
	const paddingRight = parseFloat(style.paddingRight || '0');
	const contentWidth = Math.max(0, element.clientWidth - paddingLeft - paddingRight);

	mirror.style.position = 'fixed';
	mirror.style.visibility = 'hidden';
	mirror.style.pointerEvents = 'none';
	mirror.style.top = '0px';
	mirror.style.left = '-9999px';
	mirror.style.boxSizing = 'content-box';
	mirror.style.width = `${contentWidth}px`;
	mirror.style.whiteSpace = 'pre-wrap';
	mirror.style.wordWrap = 'break-word';
	mirror.style.overflowWrap = 'break-word';

	for (const prop of PROPERTIES) {
		const cssKey = prop.replace(/[A-Z]/g, '-$&').toLowerCase();
		const val = style.getPropertyValue(cssKey);
		if (val) {
			mirror.style.setProperty(cssKey, val);
		}
	}

	const text = element.value;
	const pos = position ?? element.selectionStart;
	const index = Math.max(0, Math.min(pos, text.length));
	mirror.textContent = text.slice(0, index);

	const marker = document.createElement('span');
	marker.textContent = text.slice(index, index + 1) || '\u200b';
	mirror.appendChild(marker);

	document.body.appendChild(mirror);

	const rect = element.getBoundingClientRect();
	const markerRect = marker.getBoundingClientRect();
	const mirrorRect = mirror.getBoundingClientRect();

	const x = rect.left + (markerRect.left - mirrorRect.left) - element.scrollLeft;
	const y = rect.top + (markerRect.top - mirrorRect.top) - element.scrollTop + markerRect.height / 2;

	document.body.removeChild(mirror);

	return { x, y };
}
