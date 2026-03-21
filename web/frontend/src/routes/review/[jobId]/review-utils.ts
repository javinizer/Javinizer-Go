export interface PosterPreviewOverride {
	url: string;
	version: number;
}

export interface PosterCropBox {
	x: number;
	y: number;
	width: number;
	height: number;
}

export interface PosterCropMetrics {
	sourceWidth: number;
	sourceHeight: number;
	displayWidth: number;
	displayHeight: number;
	imageOffsetX: number;
	imageOffsetY: number;
}

export interface PosterCropState {
	xRatio: number;
	yRatio: number;
	widthRatio: number;
	heightRatio: number;
}

const LANDSCAPE_CROP_WIDTH_RATIO = 0.472;
const POSTER_TARGET_ASPECT_RATIO = 2 / 3;

export function truncatePath(path: string, maxLength: number = 80): string {
	if (path.length <= maxLength) return path;

	const ellipsis = '...';
	const charsToShow = maxLength - ellipsis.length;
	const frontChars = Math.ceil(charsToShow * 0.4);
	const backChars = Math.floor(charsToShow * 0.6);

	return path.slice(0, frontChars) + ellipsis + path.slice(-backChars);
}

export function clamp(value: number, min: number, max: number): number {
	return Math.min(max, Math.max(min, value));
}

export function normalizeCropBox(box: PosterCropBox, metrics: PosterCropMetrics): PosterCropState {
	return {
		xRatio: box.x / metrics.sourceWidth,
		yRatio: box.y / metrics.sourceHeight,
		widthRatio: box.width / metrics.sourceWidth,
		heightRatio: box.height / metrics.sourceHeight
	};
}

export function restoreCropBox(
	state: PosterCropState,
	sourceWidth: number,
	sourceHeight: number
): PosterCropBox {
	const width = clamp(Math.round(state.widthRatio * sourceWidth), 1, sourceWidth);
	const height = clamp(Math.round(state.heightRatio * sourceHeight), 1, sourceHeight);
	const maxX = Math.max(0, sourceWidth - width);
	const maxY = Math.max(0, sourceHeight - height);

	return {
		x: clamp(Math.round(state.xRatio * sourceWidth), 0, maxX),
		y: clamp(Math.round(state.yRatio * sourceHeight), 0, maxY),
		width,
		height
	};
}

export function getDefaultPosterCropBox(sourceWidth: number, sourceHeight: number): PosterCropBox {
	const sourceAspect = sourceWidth / sourceHeight;

	if (sourceAspect > 1.2) {
		const width = Math.max(1, Math.round(sourceWidth * LANDSCAPE_CROP_WIDTH_RATIO));
		return {
			x: sourceWidth - width,
			y: 0,
			width,
			height: sourceHeight
		};
	}

	let width = sourceWidth;
	let height = sourceHeight;
	if (sourceAspect > POSTER_TARGET_ASPECT_RATIO) {
		width = Math.max(1, Math.round(sourceHeight * POSTER_TARGET_ASPECT_RATIO));
	} else {
		height = Math.max(1, Math.round(sourceWidth / POSTER_TARGET_ASPECT_RATIO));
	}

	return {
		x: Math.max(0, Math.floor((sourceWidth - width) / 2)),
		y: Math.max(0, Math.floor((sourceHeight - height) / 2)),
		width,
		height
	};
}
