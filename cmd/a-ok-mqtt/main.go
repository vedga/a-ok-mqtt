package main

import (
	"fmt"
	"os"
	"time"

	"go.bug.st/serial"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	const (
		envLogLevel = `LOG_LEVEL`
		tagLogger   = `main`
	)

	if level, found := os.LookupEnv(envLogLevel); found {
		// Dynamic log level switcher (for possible future use)
		atom := zap.NewAtomicLevel()

		encoderCfg := zap.NewProductionEncoderConfig()
		// To keep the example deterministic, disable timestamps in the output.
		//		encoderCfg.TimeKey = ""
		logger := zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			atom,
		)).Named(tagLogger)

		l := atom.Level()
		_ = l.Set(level)

		atom.SetLevel(zapcore.DebugLevel)

		zap.ReplaceGlobals(logger)
	} else {
		// Production logger
		zap.ReplaceGlobals(zap.Must(zap.NewProduction()))
	}
}

func main() {
	const (
		envUartConnection = `UART_CONNECTION`
	)

	uartName, found := os.LookupEnv(envUartConnection)
	if !found {
		zap.L().Error(`Environment variable "` + envUartConnection + `" not specified`)
		return
	}

	uartMode := &serial.Mode{
		BaudRate: 2400,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		InitialStatusBits: &serial.ModemOutputBits{
			DTR: false,
			RTS: false,
		},
	}
	uart, e := serial.Open(uartName, uartMode)
	if nil != e {
		zap.L().Error(`Error open serial communication port`, zap.Error(e))
		return
	}

	go func() {
		rx, e := serial.Open(`/dev/tty.usbserial-1120`, uartMode)
		if nil != e {
			zap.L().Error(`Error open serial communication port (rx)`, zap.Error(e))
			return
		}

		defer func() {
			if e := rx.Close(); nil != e {
				zap.L().Error("Error close serial communication port (rx)", zap.Error(e))
			}

			zap.L().Debug(`UART port shutdown operation complete (rx)`)
		}()

		zap.L().Debug(`UART ready (rx)`)

		for {
			_ = uart.SetReadTimeout(5 * time.Second)
			b := make([]byte, 1)
			n, e := rx.Read(b)
			zap.L().Debug(`Rx bytes`, zap.Int(`count`, n), zap.String(`data`, fmt.Sprintf(`%b`, b)), zap.Error(e))
		}
	}()

	zap.L().Debug(`UART port opened`)

	defer func() {
		if e := uart.Close(); nil != e {
			zap.L().Error("Error close serial communication port", zap.Error(e))
		}

		zap.L().Debug(`UART port shutdown operation complete`)
	}()

	if true {
		n, e := uart.Write([]byte{
			0b00000000,
		})
		zap.L().Debug(`Sent`, zap.Int(`count`, n), zap.Error(e))
		e = uart.Drain()
		zap.L().Debug(`Sent operation complete`, zap.Int(`count`, n), zap.Error(e))

		n, e = uart.Write([]byte{
			0b10101011,
		})
		zap.L().Debug(`Sent`, zap.Int(`count`, n), zap.Error(e))
		e = uart.Drain()
		zap.L().Debug(`Sent operation complete`, zap.Int(`count`, n), zap.Error(e))
		/*
			n, e := uart.Write([]byte{
				//0x9a,
				0x00,
				0x00,
				0x00,
				0b11111111,
				0b11111111,
				0b11111111,
				0b11111111,
				//0xcc,
				//			0xcc,
			})
			zap.L().Debug(`Sent`, zap.Int(`count`, n), zap.Error(e))
			e = uart.Drain()
			zap.L().Debug(`Sent operation complete`, zap.Int(`count`, n), zap.Error(e))

		*/
	}

	/*
		_ = uart.SetReadTimeout(1 * time.Second)
		b := make([]byte, 10)
		n, e = uart.Read(b)
		zap.L().Debug(`Read bytes`, zap.Int(`count`, n), zap.Error(e))

		//	uart.Read()

	*/

	time.Sleep(2 * time.Second)
}
