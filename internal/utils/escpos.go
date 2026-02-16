package utils

import (
	"bytes"
	// Aquí podrías importar librerías de imagen si vas a procesar logos dinámicamente
)

type TicketBuilder struct {
	buffer bytes.Buffer
}

func NewTicketBuilder() *TicketBuilder {
	t := &TicketBuilder{}
	// Init printer: ESC @
	t.buffer.Write([]byte{0x1B, 0x40})
	return t
}

func (t *TicketBuilder) AddText(text string) {
	// Convertir a CP850 para que salgan las tildes si la impresora está configurada así
	// O simplemente enviar UTF-8 si es moderna.
	// Por ahora enviamos raw, asumiendo que Go maneja bien UTF-8 y la impresora también.
	// Si salen caracteres raros, usamos charmap.CodePage850.NewEncoder()

	// Opción simple:
	t.buffer.WriteString(text)
}

func (t *TicketBuilder) AddTextLn(text string) {
	t.AddText(text + "\n")
}

func (t *TicketBuilder) AlignCenter() {
	t.buffer.Write([]byte{0x1B, 0x61, 0x01})
}

func (t *TicketBuilder) AlignLeft() {
	t.buffer.Write([]byte{0x1B, 0x61, 0x00})
}

func (t *TicketBuilder) AlignRight() {
	t.buffer.Write([]byte{0x1B, 0x61, 0x02})
}

func (t *TicketBuilder) SetBold(enable bool) {
	val := byte(0x00)
	if enable {
		val = 0x01
	}
	t.buffer.Write([]byte{0x1B, 0x45, val})
}

func (t *TicketBuilder) Feed(lines int) {
	t.buffer.Write([]byte{0x1B, 0x64, byte(lines)})
}

func (t *TicketBuilder) Cut() {
	// GS V m (cortar)
	t.buffer.Write([]byte{0x1D, 0x56, 0x42, 0x00})
}

func (t *TicketBuilder) GetBytes() []byte {
	return t.buffer.Bytes()
}

// Nota sobre LOGOS:
// La mejor forma para logos estáticos en estas impresoras es:
//  1. Usar la herramienta del fabricante (3nStar Utility) para subir el logo a la memoria NV de la impresora.
//  2. Usar el comando "Imprimir logo NV" desde Go:
//     FS p n m (1C 70 01 00) -> Imprime el logo guardado en la posición 1
func (t *TicketBuilder) PrintNVLogo(logoIndex int) {
	// Comando: FS p n m
	t.buffer.Write([]byte{0x1C, 0x70, byte(logoIndex), 0x00})
}
