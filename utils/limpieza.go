package utils

import (
	"regexp"
	"strings"
)

func LimpiarTexto(text string) string {
	text = strings.ToLower(text)

	// eliminar tildes
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i",
		"ó", "o", "ú", "u", "ñ", "n",
	)
	text = replacer.Replace(text)

	// eliminar caracteres especiales
	reg := regexp.MustCompile(`[^a-z0-9\s]`)
	text = reg.ReplaceAllString(text, "")

	return text
}
