package utils

var stopwords = map[string]bool{
	"el": true, "la": true, "los": true, "las": true,
	"de": true, "del": true,
	"y": true, "o": true,
	"en": true, "para": true, "por": true,
	"un": true, "una": true,
	"con": true, "sin": true,
	"sobre": true, "entre": true,
	"que": true, "se": true,
	"a": true, "al": true,
	"lo": true, "como": true,
	"su": true, "sus": true,
	"es": true, "son": true,
}

func RemoverStopwords(tokens []string) []string {
	var resultado []string

	for _, palabra := range tokens {
		if !stopwords[palabra] {
			resultado = append(resultado, palabra)
		}
	}

	return resultado
}
