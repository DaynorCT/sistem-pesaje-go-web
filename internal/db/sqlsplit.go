package db

import "strings"

func splitStatements(sql string) []string {
	var sentencias []string
	var actual strings.Builder
	enComillaSimple := false
	enComillaDoble := false
	enAcento := false
	enComentarioLinea := false
	enComentarioBloque := false

	runes := []rune(sql)
	for i := 0; i < len(runes); i++ {
		caracter := runes[i]
		siguiente := rune(0)
		if i+1 < len(runes) {
			siguiente = runes[i+1]
		}

		if enComentarioLinea {
			if caracter == '\n' {
				enComentarioLinea = false
				actual.WriteRune(caracter)
			}
			continue
		}
		if enComentarioBloque {
			if caracter == '*' && siguiente == '/' {
				enComentarioBloque = false
				i++
			}
			continue
		}
		if !enComillaSimple && !enComillaDoble && !enAcento {
			if caracter == '-' && siguiente == '-' {
				enComentarioLinea = true
				i++
				continue
			}
			if caracter == '/' && siguiente == '*' {
				enComentarioBloque = true
				i++
				continue
			}
			if caracter == ';' {
				appendIfNonEmpty(&sentencias, &actual)
				actual.Reset()
				continue
			}
		}
		if caracter == '\'' && !enComillaDoble && !enAcento {
			if siguiente == '\'' {
				actual.WriteRune(caracter)
				actual.WriteRune(siguiente)
				i++
				continue
			}
			enComillaSimple = !enComillaSimple
		} else if caracter == '"' && !enComillaSimple && !enAcento {
			enComillaDoble = !enComillaDoble
		} else if caracter == '`' && !enComillaSimple && !enComillaDoble {
			enAcento = !enAcento
		}
		actual.WriteRune(caracter)
	}
	appendIfNonEmpty(&sentencias, &actual)
	return sentencias
}

func appendIfNonEmpty(sentencias *[]string, actual *strings.Builder) {
	s := strings.TrimSpace(actual.String())
	if s != "" {
		*sentencias = append(*sentencias, s)
	}
}
