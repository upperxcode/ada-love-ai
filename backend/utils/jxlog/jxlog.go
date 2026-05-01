package jxlog

import (
	"fmt"
	"os"
	"strings"
)

const DEBUG = true

// Cores ANSI para formatação no terminal
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func colorPrefix(level string) (string, string) {
	var color string
	var prefix string

	switch strings.ToUpper(level) {
	case "INFO":
		color = ColorGreen
		prefix = "[INFO] "
	case "WARNING":
		color = ColorYellow
		prefix = "[WARNING] "
	case "ERROR":
		color = ColorRed
		prefix = "[ERROR] "
	case "FATAL":
		color = ColorRed
		prefix = "[FATAL] "
		// os.Exit(1) // Pode ser melhor deixar o chamador decidir se quer terminar
	case "PANIC":
		color = ColorRed
		prefix = "[PANIC] "
		// panic(fmt.Sprint(str...)) // Pode ser melhor deixar o chamador decidir o que fazer
	case "CRITICAL":
		color = ColorRed
		prefix = "[CRITICAL] "
	default:
		color = ColorWhite
		prefix = "[DEBUG] " // Nível padrão
	}
	return color, prefix
}

// Função genérica de log com cores e níveis
func log(level string, str ...any) {
	if DEBUG {
		color, prefix := colorPrefix(level)

		fmt.Print(color, prefix)
		fmt.Print(str...) // fmt.Print para evitar quebras de linha extras
		fmt.Println(ColorReset)
		if strings.ToUpper(level) == "FATAL" {
			os.Exit(1)
		}
		if strings.ToUpper(level) == "PANIC" {
			panic(fmt.Sprint(str...))
		}
	}
}

// Função genérica de log com cores e níveis e formatação
func logf(level string, format string, a ...any) {
	if DEBUG {
		color, prefix := colorPrefix(level)

		fmt.Print(color, prefix)
		fmt.Printf(format, a...) // Usamos fmt.Printf para aplicar a formatação
		fmt.Println(ColorReset)

	}
	if strings.ToUpper(level) == "FATAL" {
		os.Exit(1)
	}
	if strings.ToUpper(level) == "PANIC" {
		panic(fmt.Sprint(a...))
	}
}

// Funções auxiliares para cada nível de log
func Info(str ...any) {
	log("INFO", str...)
}

func Warning(str ...any) {
	log("WARNING", str...)
}

func Error(str ...any) {
	log("ERROR", str...)
}

func Fatal(str ...any) {
	log("FATAL", str...)
}

func Panic(str ...any) {
	log("PANIC", str...)
}

func Critical(str ...any) {
	log("CRITICAL", str...)
}

func Debug(str ...any) {
	log("DEBUG", str...)
}

// Funções auxiliares para cada nível de log com formatação
func Infof(format string, a ...any) {
	logf("INFO", format, a...)
}

func Warningf(format string, a ...any) {
	logf("WARNING", format, a...)
}

func Errorf(format string, a ...any) {
	logf("ERROR", format, a...)
}

func Fatalf(format string, a ...any) {
	logf("FATAL", format, a...)
}

func Panicf(format string, a ...any) {
	logf("PANIC", format, a...)
}

func Criticalf(format string, a ...any) {
	logf("CRITICAL", format, a...)
}

func Debugf(format string, a ...any) {
	logf("DEBUG", format, a...)
}
