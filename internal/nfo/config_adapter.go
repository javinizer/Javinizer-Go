package nfo

import "github.com/javinizer/javinizer-go/internal/config"

// ConfigFromAppConfig converts application config to NFO generator config
func ConfigFromAppConfig(appCfg *config.NFOConfig, outputCfg *config.OutputConfig) *Config {
	if appCfg == nil {
		return DefaultConfig()
	}

	groupActress := false
	if outputCfg != nil {
		groupActress = outputCfg.GroupActress
	}

	return &Config{
		ActorFirstNameOrder:  appCfg.FirstNameOrder,
		ActorJapaneseNames:   appCfg.ActressLanguageJA,
		UnknownActress:       appCfg.UnknownActressText,
		NFOFilenameTemplate:  appCfg.FilenameTemplate,
		PerFile:              appCfg.PerFile,
		IncludeStreamDetails: appCfg.IncludeStreamDetails,
		IncludeFanart:        appCfg.IncludeFanart,
		IncludeTrailer:       appCfg.IncludeTrailer,
		DefaultRatingSource:  appCfg.RatingSource,
		GroupActress:         groupActress,
	}
}
