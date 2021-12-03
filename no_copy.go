// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

// Lock is a no-op used by -copylocks checker from `go vet`.
type NoCopy interface {
	Lock()
	Unlock()
}

// Lock is a no-op used by -copylocks checker from `go vet`.
type NoCopyImpl struct{}

func (*NoCopyImpl) Lock()   {}
func (*NoCopyImpl) Unlock() {}
