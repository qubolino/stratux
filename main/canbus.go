package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/brutella/can"
)

type CanData struct {
	// from Indicated Air Speed sensor
	muIAS *sync.Mutex
	IASLastTime time.Time
	IASValueKts uint16

	// from Flaps
	muFlaps *sync.Mutex
	FlapsLastTime time.Time
	FlapsValuePct uint16
}

func listenToCanBus() {
	myCanData.muIAS = &sync.Mutex{}
	myCanData.muFlaps = &sync.Mutex{}

	iface, err := net.InterfaceByName("can0")

	if err != nil {
		log.Fatalf("Could not find network interface %s (%v)", "can0", err)
	}

	conn, err := can.NewReadWriteCloserForInterface(iface)

	if err != nil {
		log.Fatal(err)
	}

	bus := can.NewBus(conn)
	bus.SubscribeFunc(processCANFrame)

	bus.ConnectAndPublish()
}

func processCANFrame(frm can.Frame) {
	if (frm.ID == 0x28) {
		ias := binary.LittleEndian.Uint16(frm.Data[:2])
		myCanData.muIAS.Lock()
		myCanData.IASValueKts = ias
		myCanData.IASLastTime = stratuxClock.Time
		myCanData.muIAS.Unlock()
		// log.Printf("received IAS: %d\n", ias)
	} else if (frm.ID == 0x60) {
		flaps := binary.LittleEndian.Uint16(frm.Data[:2])
		myCanData.muFlaps.Lock()
		myCanData.FlapsValuePct = flaps
		myCanData.FlapsLastTime = stratuxClock.Time
		myCanData.muFlaps.Unlock()
		// log.Printf("received Flaps: %d\n", flaps)
	} else {
		data := trimSuffix(frm.Data[:], 0x00)
		length := fmt.Sprintf("[%x]", frm.Length)
		log.Printf("%-3s %-4x %-3s % -24X '%s'\n", "can0", frm.ID, length, data, printableString(data[:]))
	}
	}

// trim returns a subslice of s by slicing off all trailing b bytes.
func trimSuffix(s []byte, b byte) []byte {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != b {
			return s[:i+1]
		}
	}

	return []byte{}
}

// printableString creates a string from s and replaces non-printable bytes (i.e. 0-32, 127)
// with '.' – similar how candump from can-utils does it.
func printableString(s []byte) string {
	var ascii []byte
	for _, b := range s {
		if b < 32 || b > 126 {
			b = byte('.')

		}
		ascii = append(ascii, b)
	}

	return string(ascii)
}
