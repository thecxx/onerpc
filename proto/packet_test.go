package proto

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestPacket_Write(t *testing.T) {
	sp := NewPacket()
	sp.SetSeq(rand.Uint64())
	// p.SetOption(OptOneWay, true)

	// sp.Payload = []byte(`Hello World!`)
	sp.Store([]byte(`Hello World!`))

	var buff bytes.Buffer

	sp.WriteTo(&buff)

	t.Logf("%+v", buff.Bytes())

	rp := NewPacket()

	_, err := rp.ReadFrom(&buff)
	if err != nil {
		t.Errorf("Read failed, error is %s", err.Error())
		return
	}

	t.Logf("%+v", *rp.Header)
	// t.Logf("%+v", rp.Payload)
	// t.Logf("%s", string(rp.Payload))

}
