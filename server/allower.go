package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

// Allower may be implemented to specifically allow or disallow players from
// joining a Server, by setting the specific Allower implementation through a
// call to Server.Allow.
type Allower interface {
	// Allow filters what connections are allowed to connect to the Server. The
	// address, identity data, and client data of the connection are passed. If
	// Admit returns false, the connection is closed with the string returned as
	// the disconnect message. WARNING: Use the client data at your own risk, it
	// cannot be trusted because it can be freely changed by the player
	// connecting.

	Allow(addr net.Addr, d login.IdentityData, c login.ClientData) (string, bool)
}

// allower is the standard Allower implementation. It accepts all connections.
type allower struct {
	Proxy            bool
	ForwardingSecret string
}

func signXUID(xuid string, forwardingSecret string) string {
	mac := hmac.New(sha256.New, []byte(forwardingSecret))
	mac.Write([]byte(xuid))
	return hex.EncodeToString(mac.Sum(nil))
}

// Allow always returns true.
func (a allower) Allow(conn net.Addr, d login.IdentityData, c login.ClientData) (string, bool) {
	if !a.Proxy {
		//log.Printf("Player admitted: %s (XUID: %s)", d.DisplayName, d.XUID)
		return "", true
	}
	rawsigniture := c.SkinAnimationData
	expectedSig := signXUID(d.XUID, a.ForwardingSecret)
	if rawsigniture != expectedSig {
		return "Invalid signature.", false
	}

	log.Printf("Player admitted: %s (XUID: %s)", d.DisplayName, d.XUID)
	return "", true
}
