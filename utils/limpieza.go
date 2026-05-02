package utils

import (
	"regexp"
	"strings"
)

func LimpiarTexto(text string) string {
	text = strings.ToLower(text)

	// 🔹 corregir caracteres mal codificados comunes
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a", "â", "a",
		"é", "e", "è", "e", "ë", "e", "ê", "e",
		"í", "i", "ì", "i", "ï", "i", "î", "i",
		"ó", "o", "ò", "o", "ö", "o", "ô", "o",
		"ú", "u", "ù", "u", "ü", "u", "û", "u",
		"ñ", "n",
		"�", "", // eliminar basura
	)

	text = replacer.Replace(text)

	// 🔹 eliminar caracteres no válidos
	re := regexp.MustCompile(`[^a-z\s]`)
	text = re.ReplaceAllString(text, "")

	// 🔹 limpiar espacios
	text = strings.TrimSpace(text)

	return text
}
