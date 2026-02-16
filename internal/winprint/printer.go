package winprint

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows" // Necesitas hacer go get golang.org/x/sys/windows
)

// Constantes y Structs necesarios
const (
	PRINTER_ENUM_LOCAL       = 2
	PRINTER_ENUM_CONNECTIONS = 4
)

type DOC_INFO_1 struct {
	DocName    *uint16
	OutputFile *uint16
	Datatype   *uint16
}

type PRINTER_INFO_5 struct {
	PrinterName              *uint16
	PortName                 *uint16
	Attributes               uint32
	DeviceNotSelectedTimeout uint32
	TransmissionRetryTimeout uint32
}

// --- Funciones Públicas ---

// ListLocalPrinters devuelve los nombres de las impresoras instaladas
func ListLocalPrinters() ([]string, error) {
	const flags = PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS
	var needed, returned uint32
	buf := make([]byte, 1)

	// Primera llamada para saber tamaño de buffer
	err := EnumPrinters(flags, nil, 5, &buf[0], uint32(len(buf)), &needed, &returned)
	if err != nil {
		if err != syscall.ERROR_INSUFFICIENT_BUFFER {
			return nil, err
		}
		buf = make([]byte, needed)
		// Segunda llamada con buffer correcto
		err = EnumPrinters(flags, nil, 5, &buf[0], uint32(len(buf)), &needed, &returned)
		if err != nil {
			return nil, err
		}
	}

	ps := (*[1024]PRINTER_INFO_5)(unsafe.Pointer(&buf[0]))[:returned:returned]
	names := make([]string, 0, returned)
	for _, p := range ps {
		names = append(names, windows.UTF16PtrToString(p.PrinterName))
	}
	return names, nil
}

// SendBytesToPrinter envía datos RAW (ESC/POS) a una impresora por nombre
func SendBytesToPrinter(printerName string, data []byte) error {
	var hPrinter syscall.Handle

	// 1. Abrir Impresora
	docName := "Go Print Job"
	dataType := "RAW"

	ptrName, _ := syscall.UTF16PtrFromString(printerName)
	ptrDocName, _ := syscall.UTF16PtrFromString(docName)
	ptrDataType, _ := syscall.UTF16PtrFromString(dataType)

	if err := OpenPrinter(ptrName, &hPrinter, 0); err != nil {
		return err
	}
	defer ClosePrinter(hPrinter)

	// 2. Iniciar Documento
	di := DOC_INFO_1{
		DocName:    ptrDocName,
		OutputFile: nil,
		Datatype:   ptrDataType,
	}

	if err := StartDocPrinter(hPrinter, 1, &di); err != nil {
		return err
	}
	defer EndDocPrinter(hPrinter)

	// 3. Iniciar Página
	if err := StartPagePrinter(hPrinter); err != nil {
		return err
	}
	defer EndPagePrinter(hPrinter)

	// 4. Escribir Bytes
	var written uint32
	if err := WritePrinter(hPrinter, &data[0], uint32(len(data)), &written); err != nil {
		return err
	}

	return nil
}
