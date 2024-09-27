package normalize

import sluglib "github.com/gosimple/slug"

func Email(email string) string {
	sluglib.MaxLength = 63
	sluglib.CustomSub = map[string]string{
		"_": "-",
		"@": "-at-",
	}

	return sluglib.Make(email)
}
