package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger define la interfaz para el registro thread-safe de salidas de los procesos.
type Logger interface {
	LogStdout(processName string, line string)
	LogStderr(processName string, line string)
	LogInfo(processName string, message string)
	LogError(processName string, message string)
}

// StreamType identifica el tipo de flujo de salida.
type StreamType string

const (
	StreamStdout StreamType = "STDOUT"
	StreamStderr StreamType = "STDERR"
	StreamInfo   StreamType = "INFO"
	StreamError  StreamType = "ERROR"
)

// ProcessLogger implementa un formateador thread-safe de logs etiquetados por proceso.
type ProcessLogger struct {
	mu         sync.Mutex
	out        io.Writer
	timeFormat string
}

// NewProcessLogger crea un nuevo Logger configurado por defecto a os.Stdout.
func NewProcessLogger() *ProcessLogger {
	return &ProcessLogger{
		out:        os.Stdout,
		timeFormat: "15:04:05",
	}
}

// NewProcessLoggerWithWriter crea un Logger utilizando un io.Writer personalizado (útil para pruebas).
func NewProcessLoggerWithWriter(w io.Writer) *ProcessLogger {
	return &ProcessLogger{
		out:        w,
		timeFormat: "15:04:05",
	}
}

func (l *ProcessLogger) log(processName string, stream StreamType, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(l.timeFormat)
	fmt.Fprintf(l.out, "[%s] [%s] [%s] %s\n", timestamp, processName, stream, message)
}

func (l *ProcessLogger) LogStdout(processName string, line string) {
	l.log(processName, StreamStdout, line)
}

func (l *ProcessLogger) LogStderr(processName string, line string) {
	l.log(processName, StreamStderr, line)
}

func (l *ProcessLogger) LogInfo(processName string, message string) {
	l.log(processName, StreamInfo, message)
}

func (l *ProcessLogger) LogError(processName string, message string) {
	l.log(processName, StreamError, message)
}

// StreamPipe lee líneas de un io.Reader y las redirige al logger usando la función de log indicada.
func StreamPipe(r io.Reader, processName string, logFunc func(processName string, line string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logFunc(processName, scanner.Text())
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		logFunc(processName, fmt.Sprintf("error en lectura de flujo: %v", err))
	}
}
