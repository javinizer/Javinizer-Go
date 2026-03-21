import type { BatchJobResponse, Movie } from '$lib/api/types';

interface ReviewPageControllerDeps {
	getJob: () => BatchJobResponse | null;
	getCurrentMovie: () => Movie | null;
	getMovieResultsLength: () => number;
	getCurrentMovieIndex: () => number;
	setCurrentMovieIndex: (index: number) => void;
	getEditedMovies: () => Map<string, Movie>;
	getDestinationPath: () => string;
	setDestinationPath: (destinationPath: string) => void;
	getTempDestinationPath: () => string;
	setTempDestinationPath: (destinationPath: string) => void;
	setShowDestinationBrowser: (show: boolean) => void;
	setShowImageViewer: (show: boolean) => void;
	setImageViewerImages: (images: string[]) => void;
	setImageViewerIndex: (index: number) => void;
	setImageViewerTitle: (title: string | undefined) => void;
	refetchJob: () => Promise<void>;
	toastSuccess: (message: string, duration?: number) => void;
	toastError: (message: string, duration?: number) => void;
	navigateBatch: () => void | Promise<void>;
	api: {
		excludeBatchMovie: (jobId: string, movieId: string) => Promise<unknown>;
		getPreviewImageURL: (url: string) => string;
	};
}

export function createReviewPageController(deps: ReviewPageControllerDeps) {
	function hasChanges(filePath: string): boolean {
		return deps.getEditedMovies().has(filePath);
	}

	function openDestinationBrowser() {
		deps.setTempDestinationPath(deps.getDestinationPath());
		deps.setShowDestinationBrowser(true);
	}

	function confirmDestination() {
		deps.setDestinationPath(deps.getTempDestinationPath());
		deps.setShowDestinationBrowser(false);
	}

	function cancelDestination() {
		deps.setShowDestinationBrowser(false);
	}

	function previewImageURL(url: string | undefined): string {
		if (!url) return '';
		if (url.startsWith('/api/v1/')) return url;
		if (url.startsWith('/')) return url;
		return deps.api.getPreviewImageURL(url);
	}

	function openScreenshotViewer(index: number) {
		const currentMovie = deps.getCurrentMovie();
		if (!currentMovie?.screenshot_urls) return;
		deps.setImageViewerImages(currentMovie.screenshot_urls.map((url) => previewImageURL(url)));
		deps.setImageViewerIndex(index);
		deps.setImageViewerTitle(undefined);
		deps.setShowImageViewer(true);
	}

	function openCoverViewer() {
		const currentMovie = deps.getCurrentMovie();
		if (!currentMovie?.cover_url) return;
		deps.setImageViewerImages([previewImageURL(currentMovie.cover_url)]);
		deps.setImageViewerIndex(0);
		deps.setImageViewerTitle('Cover/Fanart');
		deps.setShowImageViewer(true);
	}

	function closeImageViewer() {
		deps.setShowImageViewer(false);
	}

	async function excludeCurrentMovie() {
		const currentMovie = deps.getCurrentMovie();
		const job = deps.getJob();
		if (!currentMovie || !job) return;

		try {
			await deps.api.excludeBatchMovie(job.id, currentMovie.id);
			deps.toastSuccess(`Movie ${currentMovie.id} excluded from organization`);
			await deps.refetchJob();

			const movieResultsLength = deps.getMovieResultsLength();
			if (movieResultsLength === 0) {
				await deps.navigateBatch();
				return;
			}

			if (deps.getCurrentMovieIndex() >= movieResultsLength) {
				deps.setCurrentMovieIndex(movieResultsLength - 1);
			}
		} catch (err) {
			deps.toastError(`Failed to exclude movie: ${err}`);
		}
	}

	return {
		hasChanges,
		openDestinationBrowser,
		confirmDestination,
		cancelDestination,
		previewImageURL,
		openScreenshotViewer,
		openCoverViewer,
		closeImageViewer,
		excludeCurrentMovie
	};
}
