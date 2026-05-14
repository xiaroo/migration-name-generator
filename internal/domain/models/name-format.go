package models

type NameFormat string

const (
	NameFormatCamelCase          NameFormat = "camel_case"           // userName
	NameFormatPascalCase         NameFormat = "pascal_case"          // UserName
	NameFormatSnakeCase          NameFormat = "snake_case"           // user_name
	NameFormatScreamingSnakeCase NameFormat = "screaming_snake_case" // USER_NAME
	NameFormatKebabCase          NameFormat = "kebab_case"           // user-name
	NameFormatScreamingKebabCase NameFormat = "screaming_kebab_case" // USER-NAME
	NameFormatTrainCase          NameFormat = "train_case"           // User-Name
	NameFormatDotCase            NameFormat = "dot_case"             // user.name
	NameFormatPathCase           NameFormat = "path_case"            // user/name
	NameFormatFlatCase           NameFormat = "flat_case"            // username
	NameFormatUpperFlatCase      NameFormat = "upper_flat_case"      // USERNAME
	NameFormatTitleCase          NameFormat = "title_case"           // User Name
	NameFormatSentenceCase       NameFormat = "sentence_case"        // User name
	NameFormatLowerCase          NameFormat = "lower_case"           // user name
	NameFormatUpperCase          NameFormat = "upper_case"           // USER NAME
)

func ValidNameFormats() []NameFormat {
	return []NameFormat{
		NameFormatCamelCase,
		NameFormatPascalCase,
		NameFormatSnakeCase,
		NameFormatScreamingSnakeCase,
		NameFormatKebabCase,
		NameFormatScreamingKebabCase,
		NameFormatTrainCase,
		NameFormatDotCase,
		NameFormatPathCase,
		NameFormatFlatCase,
		NameFormatUpperFlatCase,
		NameFormatTitleCase,
		NameFormatSentenceCase,
		NameFormatLowerCase,
		NameFormatUpperCase,
	}
}

func (format NameFormat) IsValid() bool {
	for _, valid := range ValidNameFormats() {
		if format == valid {
			return true
		}
	}
	return false
}

func (format NameFormat) String() string {
	return string(format)
}
