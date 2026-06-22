package session

import (
	"github.com/df-mc/dragonfly/server/player/ddui"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// DataDrivenScreenClosedHandler handles the ServerBoundDataDrivenScreenClosed packet.
type DataDrivenScreenClosedHandler struct {
	h *DDUIFormHandler
}

// Handle ...
func (d *DataDrivenScreenClosedHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, _ Controllable) error {
	pk := p.(*packet.ServerBoundDataDrivenScreenClosed)

	formID, hasID := pk.FormID.Value()

	d.h.mu.Lock()
	var (
		af     *activeDDUIForm
		active []*activeDDUIForm
	)
	if hasID {
		for _, f := range d.h.forms {
			if f.formID == formID {
				af = f
				delete(d.h.forms, f.instanceID)
				break
			}
		}
	} else {
		active = make([]*activeDDUIForm, 0, len(d.h.forms))
		for _, f := range d.h.forms {
			active = append(active, f)
		}
		d.h.forms = make(map[uint32]*activeDDUIForm)
	}

	d.h.mu.Unlock()

	reason := closeReasonToDDUI(pk.CloseReason)
	if hasID {
		if af == nil {
			return nil
		}
		af.form.OnClose(reason)
		s.writePacket(&packet.ClientBoundDataDrivenUICloseScreen{
			FormID: protocol.Option(af.formID),
		})
		sendDataStoreCleanup(s, af)
		return nil
	}

	if len(active) == 0 {
		return nil
	}
	for _, af := range active {
		af.form.OnClose(reason)
		sendDataStoreCleanup(s, af)
	}
	return nil
}

func closeReasonToDDUI(reason string) int {
	switch reason {
	case packet.DataDrivenScreenCloseReasonProgrammaticClose:
		return ddui.CloseReasonProgrammatic
	case packet.DataDrivenScreenCloseReasonProgrammaticCloseAll:
		return ddui.CloseReasonProgrammaticAll
	case packet.DataDrivenScreenCloseReasonClientCanceled:
		return ddui.Closed
	case packet.DataDrivenScreenCloseReasonUserBusy:
		return ddui.Busy
	case packet.DataDrivenScreenCloseReasonInvalidForm:
		return ddui.Invalid
	default:
		return ddui.Closed
	}
}
