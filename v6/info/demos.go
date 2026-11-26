package info

import (
	"NintendoChannel/common"
	"bytes"
	"encoding/binary"
)

type DemoTable struct {
	ID       uint32
	Title    [31]uint16
	Subtitle [31]uint16
}

func (i *Info) WriteDemos(buffer *bytes.Buffer, demos []DemoTable) {
	if len(demos) == 0 {
		return
	}
	i.Header.NumberOfDemosTables = uint32(len(demos))
	i.Header.DemosTableOffset = i.GetCurrentSize(buffer)
	err := binary.Write(buffer, binary.BigEndian, demos)
	common.CheckError(err)
}
