package netconf

import (
	"context"
	"strconv"
)

// GlobalCounter keeps a running count of every NETCONF RPC. It is incremented
// by the DefaultXMLAttr and DefaultRPCMethodWrapper functions.
//
// GlobalCounter is safe for client applications to access, use, and increment.
var GlobalCounter = NewUintCounterContext(context.Background())

// Uint is a 64-bit unsigned integer variable that satisfies the expvar.Var interface.
type Uint struct {
	readChan chan uint64
	setChan  chan uint64
	addChan  chan uint64
	val      uint64
}

// NewUintCounterContext allocates the new unsigned integer counter
// with resources that can only be cancelled by a context.
func NewUintCounterContext(ctx context.Context) *Uint {

	var u Uint

	u.readChan = make(chan uint64)
	u.setChan = make(chan uint64)
	u.addChan = make(chan uint64)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case u.readChan <- u.val:
			case val := <-u.setChan:
				u.val = val
			case delta := <-u.addChan:
				u.val += delta
			}
		}
	}()

	return &u
}

// Value returns the current value of the underlying uint64.
func (v *Uint) Value() uint64 {
	return <-v.readChan
}

// String converts the underlying uint64 to its base 10 string representation.
func (v *Uint) String() string {
	return strconv.FormatUint(<-v.readChan, 10)
}

// Add add the given delta argument to the underlying uint64 value.
func (v *Uint) Add(delta uint64) {
	v.addChan <- delta
}

// Set assigns the given value argument to the underlying uint64.
func (v *Uint) Set(value uint64) {
	v.setChan <- value
}
