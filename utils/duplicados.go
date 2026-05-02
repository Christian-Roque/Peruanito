package utils

func RemoverDuplicados(tokens []string) []string {
	vistos := make(map[string]bool)
	var resultado []string

	for _, token := range tokens {
		if !vistos[token] {
			vistos[token] = true
			resultado = append(resultado, token)
		}
	}

	return resultado
}
