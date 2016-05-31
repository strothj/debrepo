package debrepo

import (
	"net/http"
	"sync"

	"golang.org/x/crypto/openpgp"
)

// A Client is a Debian Repository client.
type Client struct {
	mu sync.Mutex
	// sources SourceList
	client  *http.Client
	keyring openpgp.KeyRing
}

// func NewClient(source SourceList, keyring openpgp.KeyRing, client *http.Client) *Client {
// 	panic("Not Implemented")
// }
