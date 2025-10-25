package database

import (
	"fmt"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the GORM database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "sqlite", "":
		dialector = sqlite.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	// Configure logger level
	logLevel := logger.Silent
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logger.Info
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DB{db}, nil
}

// AutoMigrate runs database migrations
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&models.Movie{},
		&models.MovieTranslation{},
		&models.Actress{},
		&models.Genre{},
		&models.GenreReplacement{},
		&models.History{},
	)
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// MovieRepository provides database operations for movies
type MovieRepository struct {
	db *DB
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(db *DB) *MovieRepository {
	return &MovieRepository{db: db}
}

// Create saves a new movie to the database
func (r *MovieRepository) Create(movie *models.Movie) error {
	return r.db.Create(movie).Error
}

// Update updates an existing movie
func (r *MovieRepository) Update(movie *models.Movie) error {
	return r.db.Save(movie).Error
}

// Upsert creates a new movie or updates if it already exists (by ID)
func (r *MovieRepository) Upsert(movie *models.Movie) error {
	// Check if movie exists
	existing, err := r.FindByID(movie.ID)
	if err != nil {
		// Movie doesn't exist, create it (with translations)
		return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(movie).Error
	}

	// Movie exists, update it
	// Preserve the CreatedAt timestamp from the existing record
	movie.CreatedAt = existing.CreatedAt

	// Set the MovieID for all translations
	for i := range movie.Translations {
		movie.Translations[i].MovieID = movie.ID
	}

	// Use Save which will update the record and set UpdatedAt automatically
	// This will also update associations (actresses, genres, translations)
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(movie).Error
}

// FindByID finds a movie by its ID
func (r *MovieRepository) FindByID(id string) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.Preload("Actresses").Preload("Genres").Preload("Translations").First(&movie, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// FindByContentID finds a movie by its content ID
func (r *MovieRepository) FindByContentID(contentID string) (*models.Movie, error) {
	var movie models.Movie
	err := r.db.Preload("Actresses").Preload("Genres").First(&movie, "content_id = ?", contentID).Error
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

// Delete deletes a movie by ID
func (r *MovieRepository) Delete(id string) error {
	return r.db.Delete(&models.Movie{}, "id = ?", id).Error
}

// List returns a paginated list of movies
func (r *MovieRepository) List(limit, offset int) ([]models.Movie, error) {
	var movies []models.Movie
	err := r.db.Preload("Actresses").Preload("Genres").Limit(limit).Offset(offset).Find(&movies).Error
	return movies, err
}

// ActressRepository provides database operations for actresses
type ActressRepository struct {
	db *DB
}

// NewActressRepository creates a new actress repository
func NewActressRepository(db *DB) *ActressRepository {
	return &ActressRepository{db: db}
}

// Create saves a new actress to the database
func (r *ActressRepository) Create(actress *models.Actress) error {
	return r.db.Create(actress).Error
}

// Update updates an existing actress
func (r *ActressRepository) Update(actress *models.Actress) error {
	return r.db.Save(actress).Error
}

// FindByJapaneseName finds an actress by Japanese name
func (r *ActressRepository) FindByJapaneseName(name string) (*models.Actress, error) {
	var actress models.Actress
	err := r.db.First(&actress, "japanese_name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &actress, nil
}

// FindOrCreate finds an actress or creates a new one
func (r *ActressRepository) FindOrCreate(actress *models.Actress) error {
	// Try to find by Japanese name first
	if actress.JapaneseName != "" {
		existing, err := r.FindByJapaneseName(actress.JapaneseName)
		if err == nil {
			*actress = *existing
			return nil
		}
	}

	// If not found, create new
	return r.Create(actress)
}

// List returns a paginated list of actresses
func (r *ActressRepository) List(limit, offset int) ([]models.Actress, error) {
	var actresses []models.Actress
	err := r.db.Limit(limit).Offset(offset).Find(&actresses).Error
	return actresses, err
}

// MovieTranslationRepository provides database operations for movie translations
type MovieTranslationRepository struct {
	db *DB
}

// NewMovieTranslationRepository creates a new movie translation repository
func NewMovieTranslationRepository(db *DB) *MovieTranslationRepository {
	return &MovieTranslationRepository{db: db}
}

// Upsert creates a new translation or updates if it already exists (by MovieID + Language)
func (r *MovieTranslationRepository) Upsert(translation *models.MovieTranslation) error {
	// Try to find existing translation for this movie and language
	existing, err := r.FindByMovieAndLanguage(translation.MovieID, translation.Language)
	if err != nil {
		// Translation doesn't exist, create it
		return r.db.Create(translation).Error
	}

	// Translation exists, update it
	translation.ID = existing.ID
	translation.CreatedAt = existing.CreatedAt
	return r.db.Save(translation).Error
}

// FindByMovieAndLanguage finds a translation for a specific movie and language
func (r *MovieTranslationRepository) FindByMovieAndLanguage(movieID, language string) (*models.MovieTranslation, error) {
	var translation models.MovieTranslation
	err := r.db.First(&translation, "movie_id = ? AND language = ?", movieID, language).Error
	if err != nil {
		return nil, err
	}
	return &translation, nil
}

// FindAllByMovie finds all translations for a specific movie
func (r *MovieTranslationRepository) FindAllByMovie(movieID string) ([]models.MovieTranslation, error) {
	var translations []models.MovieTranslation
	err := r.db.Where("movie_id = ?", movieID).Find(&translations).Error
	return translations, err
}

// Delete deletes a translation
func (r *MovieTranslationRepository) Delete(movieID, language string) error {
	return r.db.Delete(&models.MovieTranslation{}, "movie_id = ? AND language = ?", movieID, language).Error
}

// GenreRepository provides database operations for genres
type GenreRepository struct {
	db *DB
}

// NewGenreRepository creates a new genre repository
func NewGenreRepository(db *DB) *GenreRepository {
	return &GenreRepository{db: db}
}

// FindOrCreate finds a genre or creates a new one
func (r *GenreRepository) FindOrCreate(name string) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.FirstOrCreate(&genre, models.Genre{Name: name}).Error
	return &genre, err
}

// List returns all genres
func (r *GenreRepository) List() ([]models.Genre, error) {
	var genres []models.Genre
	err := r.db.Find(&genres).Error
	return genres, err
}

// GenreReplacementRepository provides database operations for genre replacements
type GenreReplacementRepository struct {
	db *DB
}

// NewGenreReplacementRepository creates a new genre replacement repository
func NewGenreReplacementRepository(db *DB) *GenreReplacementRepository {
	return &GenreReplacementRepository{db: db}
}

// Create adds a new genre replacement
func (r *GenreReplacementRepository) Create(replacement *models.GenreReplacement) error {
	return r.db.Create(replacement).Error
}

// Upsert creates or updates a genre replacement
func (r *GenreReplacementRepository) Upsert(replacement *models.GenreReplacement) error {
	existing, err := r.FindByOriginal(replacement.Original)
	if err != nil {
		// Doesn't exist, create it
		return r.Create(replacement)
	}

	// Exists, update it
	replacement.ID = existing.ID
	replacement.CreatedAt = existing.CreatedAt
	return r.db.Save(replacement).Error
}

// FindByOriginal finds a replacement by original genre name
func (r *GenreReplacementRepository) FindByOriginal(original string) (*models.GenreReplacement, error) {
	var replacement models.GenreReplacement
	err := r.db.First(&replacement, "original = ?", original).Error
	if err != nil {
		return nil, err
	}
	return &replacement, nil
}

// List returns all genre replacements
func (r *GenreReplacementRepository) List() ([]models.GenreReplacement, error) {
	var replacements []models.GenreReplacement
	err := r.db.Find(&replacements).Error
	return replacements, err
}

// Delete removes a genre replacement
func (r *GenreReplacementRepository) Delete(original string) error {
	return r.db.Delete(&models.GenreReplacement{}, "original = ?", original).Error
}

// GetReplacementMap returns all replacements as a map[original]replacement
func (r *GenreReplacementRepository) GetReplacementMap() (map[string]string, error) {
	replacements, err := r.List()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, r := range replacements {
		result[r.Original] = r.Replacement
	}
	return result, nil
}

// HistoryRepository provides database operations for operation history
type HistoryRepository struct {
	db *DB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// Create adds a new history record
func (r *HistoryRepository) Create(history *models.History) error {
	return r.db.Create(history).Error
}

// FindByID finds a history record by ID
func (r *HistoryRepository) FindByID(id uint) (*models.History, error) {
	var history models.History
	err := r.db.First(&history, id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// FindByMovieID finds all history records for a specific movie
func (r *HistoryRepository) FindByMovieID(movieID string) ([]models.History, error) {
	var history []models.History
	err := r.db.Where("movie_id = ?", movieID).Order("created_at DESC").Find(&history).Error
	return history, err
}

// FindByOperation finds all history records for a specific operation type
func (r *HistoryRepository) FindByOperation(operation string, limit int) ([]models.History, error) {
	var history []models.History
	query := r.db.Where("operation = ?", operation).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}

// FindByStatus finds all history records with a specific status
func (r *HistoryRepository) FindByStatus(status string, limit int) ([]models.History, error) {
	var history []models.History
	query := r.db.Where("status = ?", status).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}

// FindRecent finds the most recent history records
func (r *HistoryRepository) FindRecent(limit int) ([]models.History, error) {
	var history []models.History
	err := r.db.Order("created_at DESC").Limit(limit).Find(&history).Error
	return history, err
}

// FindByDateRange finds history records within a date range
func (r *HistoryRepository) FindByDateRange(start, end time.Time) ([]models.History, error) {
	var history []models.History
	err := r.db.Where("created_at BETWEEN ? AND ?", start, end).Order("created_at DESC").Find(&history).Error
	return history, err
}

// Count returns the total number of history records
func (r *HistoryRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.History{}).Count(&count).Error
	return count, err
}

// CountByStatus returns the count of records with a specific status
func (r *HistoryRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.History{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountByOperation returns the count of records for a specific operation
func (r *HistoryRepository) CountByOperation(operation string) (int64, error) {
	var count int64
	err := r.db.Model(&models.History{}).Where("operation = ?", operation).Count(&count).Error
	return count, err
}

// Delete removes a history record
func (r *HistoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.History{}, id).Error
}

// DeleteByMovieID removes all history records for a specific movie
func (r *HistoryRepository) DeleteByMovieID(movieID string) error {
	return r.db.Where("movie_id = ?", movieID).Delete(&models.History{}).Error
}

// DeleteOlderThan removes history records older than the specified date
func (r *HistoryRepository) DeleteOlderThan(date time.Time) error {
	return r.db.Where("created_at < ?", date).Delete(&models.History{}).Error
}

// List returns a paginated list of history records
func (r *HistoryRepository) List(limit, offset int) ([]models.History, error) {
	var history []models.History
	err := r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&history).Error
	return history, err
}
