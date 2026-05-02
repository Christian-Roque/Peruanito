package utils

var stopwords = map[string]bool{
	"de": true, "la": true, "el": true, "en": true,
	"y": true, "a": true, "los": true, "del": true,
	"se": true, "las": true, "por": true, "un": true,
	"para": true, "con": true, "no": true, "una": true,
}

func RemoverStopwords(tokens []string) []string {
	var resultado []string

	for _, t := range tokens {
		if !stopwords[t] {
			resultado = append(resultado, t)
		}
	}

	return resultado
}
