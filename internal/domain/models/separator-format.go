package models

type SeparatorFormat string

const (
	SeparatorFormatUnderscore SeparatorFormat = "underscore" // _
	SeparatorFormatDash       SeparatorFormat = "dash"       // -
	SeparatorFormatSlash      SeparatorFormat = "slash"      // /
	SeparatorFormatDot        SeparatorFormat = "dot"        // .
)

func ValidSeparatorFormats() []SeparatorFormat {
	return []SeparatorFormat{
		SeparatorFormatUnderscore,
		SeparatorFormatDash,
		SeparatorFormatSlash,
		SeparatorFormatDot,
	}
}

func (format SeparatorFormat) IsValid() bool {
	for _, valid := range ValidSeparatorFormats() {
		if format == valid {
			return true
		}
	}
	return false
}

func (format SeparatorFormat) String() string {
	return string(format)
}
