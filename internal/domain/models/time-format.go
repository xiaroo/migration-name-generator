package models

type TimeFormat string

const (
	TimeFormatYYYYMMDDHHMMSS               TimeFormat = "YYYYMMDDHHMMSS"               // 20260514090705
	TimeFormatYYYYMMDD_HHMMSS              TimeFormat = "YYYYMMDD_HHMMSS"              // 20260514_090705
	TimeFormatYYYY_MM_DD_HH_MM_SS          TimeFormat = "YYYY_MM_DD_HH_MM_SS"          // 2026_05_14_09_07_05
	TimeFormatYYYYDashMMDashDD_HHMMSS      TimeFormat = "YYYYDashMMDashDD_HHMMSS"      // 2026-05-14 090705
	TimeFormatYYYYDashMMDashDD_HH_MM_SS    TimeFormat = "YYYYDashMMDashDD_HH_MM_SS"    // 2026-05-14 09:07:05
	TimeFormatYYYYDashMMDashDDTHHMMSS      TimeFormat = "YYYYDashMMDashDDTHHMMSS"      // 2026-05-14T090705
	TimeFormatYYYYDashMMDashDDTHH_MM_SS    TimeFormat = "YYYYDashMMDashDDTHH_MM_SS"    // 2026-05-14T09:07:05
	TimeFormatYYYYSlashMMSlashDD_HH_MM_SS  TimeFormat = "YYYYSlashMMSlashDD_HH_MM_SS"  // 2026/05/14 09:07:05
	TimeFormatYYYYDotMMDotDD_HH_MM_SS      TimeFormat = "YYYYDotMMDotDD_HH_MM_SS"      // 2026.05.14 09:07:05
	TimeFormatDDMMYYYYHHMMSS               TimeFormat = "DDMMYYYYHHMMSS"               // 14052026090705
	TimeFormatDD_MM_YYYY_HH_MM_SS          TimeFormat = "DD_MM_YYYY_HH_MM_SS"          // 14_05_2026_09_07_05
	TimeFormatDDDashMMDashYYYY_HH_MM_SS    TimeFormat = "DDDashMMDashYYYY_HH_MM_SS"    // 14-05-2026 09:07:05
	TimeFormatDDSplashMMSlashYYYY_HH_MM_SS TimeFormat = "DDSplashMMSlashYYYY_HH_MM_SS" // 14/05/2026 09:07:05
	TimeFormatDDDotMMDotYYYY_HH_MM_SS      TimeFormat = "DDDotMMDotYYYY_HH_MM_SS"      // 14.05.2026 09:07:05
	TimeFormatMMDDYYYYHHMMSS               TimeFormat = "MMDDYYYYHHMMSS"               // 05142026090705
	TimeFormatMM_DD_YYYY_HH_MM_SS          TimeFormat = "MM_DD_YYYY_HH_MM_SS"          // 05_14_2026_09_07_05
	TimeFormatMMDashDDDashYYYY_HH_MM_SS    TimeFormat = "MMDashDDDashYYYY_HH_MM_SS"    // 05-14-2026 09:07:05
	TimeFormatMMSlashDDSlashYYYY_HH_MM_SS  TimeFormat = "MMSlashDDSlashYYYY_HH_MM_SS"  // 05/14/2026 09:07:05
	TimeFormatFilenameSafe                 TimeFormat = "FilenameSafe"                 // 2026-05-14_09-07-05
	TimeFormatISO8601                      TimeFormat = "ISO8601"                      // 2026-05-14T09:07:05
	TimeFormatRFC3339                      TimeFormat = "RFC3339"                      // 2026-05-14T09:07:05+07:00
)

func ValidTimeFormats() []TimeFormat {
	return []TimeFormat{
		TimeFormatYYYYMMDDHHMMSS,
		TimeFormatYYYYMMDD_HHMMSS,
		TimeFormatYYYY_MM_DD_HH_MM_SS,
		TimeFormatYYYYDashMMDashDD_HHMMSS,
		TimeFormatYYYYDashMMDashDD_HH_MM_SS,
		TimeFormatYYYYDashMMDashDDTHHMMSS,
		TimeFormatYYYYDashMMDashDDTHH_MM_SS,
		TimeFormatYYYYSlashMMSlashDD_HH_MM_SS,
		TimeFormatYYYYDotMMDotDD_HH_MM_SS,
		TimeFormatDDMMYYYYHHMMSS,
		TimeFormatDD_MM_YYYY_HH_MM_SS,
		TimeFormatDDDashMMDashYYYY_HH_MM_SS,
		TimeFormatDDSplashMMSlashYYYY_HH_MM_SS,
		TimeFormatDDDotMMDotYYYY_HH_MM_SS,
		TimeFormatMMDDYYYYHHMMSS,
		TimeFormatMM_DD_YYYY_HH_MM_SS,
		TimeFormatMMDashDDDashYYYY_HH_MM_SS,
		TimeFormatMMSlashDDSlashYYYY_HH_MM_SS,
		TimeFormatFilenameSafe,
		TimeFormatISO8601,
		TimeFormatRFC3339,
	}
}

func (format TimeFormat) IsValid() bool {
	for _, valid := range ValidTimeFormats() {
		if format == valid {
			return true
		}
	}
	return false
}

func (format TimeFormat) String() string {
	return string(format)
}
