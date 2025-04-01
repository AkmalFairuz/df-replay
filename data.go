package replay

import (
	"bytes"
	"fmt"
	"github.com/akmalfairuz/df-replay/action"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"io"
)

type Data struct {
	id      uuid.UUID
	actions map[uint32][]action.Action
}

func NewData(id uuid.UUID) *Data {
	return &Data{
		id: id,
	}
}

func (d *Data) LoadActions(r io.Reader) error {
	decoder, err := zstd.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create zstd reader: %w", err)
	}
	defer decoder.Close()

	buffer := bytes.NewBuffer(make([]byte, 0, 8192)) // 8KB
	_, err = io.Copy(buffer, decoder)
	if err != nil {
		return fmt.Errorf("failed to copy decompressed data: %w", err)
	}

	dec := protocol.NewReader(buffer, 0, false)
	var tickLen uint32
	dec.Uint32(&tickLen)
	d.actions = make(map[uint32][]action.Action, tickLen)
	for i := uint32(0); i < tickLen; i++ {
		var tick uint32
		dec.Varuint32(&tick)
		var actionLen uint32
		dec.Varuint32(&actionLen)
		actions := make([]action.Action, actionLen)
		for j := uint32(0); j < actionLen; j++ {
			var id uint8
			dec.Uint8(&id)
			var act action.Action
			action.Read(dec, &act)
			actions[j] = act
		}
		d.actions[tick] = actions
	}
	return nil
}
