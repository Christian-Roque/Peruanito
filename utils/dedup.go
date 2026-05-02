package utils

// 🔹 Elimina tokens repetidos manteniendo el orden original
func RemoverDuplicadosTokens(tokens []string) []string {
	vistos := make(map[string]bool)
	resultado := make([]string, 0, len(tokens))

	for _, t := range tokens {
		if !vistos[t] {
			vistos[t] = true
			resultado = append(resultado, t)
		}
	}

	return resultado
}
