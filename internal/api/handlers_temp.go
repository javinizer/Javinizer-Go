package api

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// serveTempPoster serves temporarily cropped posters from data/temp/posters/
// These are created during batch scraping for preview in the review page
func serveTempPoster() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("jobId")
		movieID := c.Param("movieId")

		// Construct path: data/temp/posters/{jobId}/{movieId}.jpg
		posterPath := filepath.Join("data", "temp", "posters", jobID, movieID+".jpg")

		// Serve the file
		c.File(posterPath)
	}
}
