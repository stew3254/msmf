package utils

// RingBuffer is a circular buffer interface
type RingBuffer interface {
	Get() (interface{}, bool)
	Pop() (interface{}, bool)
	Push(e interface{}) bool
	Clear()
	Len() int
	Cap() int
	SetCap(i int) bool
	IsEmpty() bool
	IsFull() bool
}

// BytesRing is a RingBuffer implementation for storing []byte
type BytesRing struct {
	// Position where the next insertion will occur. This is publicly accessible on purpose
	Index int
	// Length of the buffer
	len int
	// Capacity of the buffer
	cap int
	// The buffer where data presides. This is publicly accessible on purpose
	Buffer [][]byte
}

// NewBytesRing creates a new BytesRing with the given capacity
func NewBytesRing(cap int) *BytesRing {
	return &BytesRing{
		Index:  0,
		len:    0,
		cap:    cap,
		Buffer: make([][]byte, cap, cap),
	}
}

// Get the first value from the ring buffer
func (b *BytesRing) Get() ([]byte, bool) {
	if b.IsEmpty() {
		// Could not get an element
		return nil, false
	}

	// Return the first thing
	return b.Buffer[b.Index], true
}

// Pop the first value from the ring buffer
func (b *BytesRing) Pop() ([]byte, bool) {
	if b.IsEmpty() {
		// Could not get an element
		return nil, false
	}

	e := b.Buffer[b.Index]
	// Increase the index
	b.Index = (b.Index + 1) % b.cap
	// Subtract 1 from the length
	b.len -= 1
	return e, true

}

// Push adds elements onto the end of the ring
func (b *BytesRing) Push(e []byte) bool {
	// Cannot add the element because the buffer is full
	if b.IsFull() {
		return false
	}

	// Set the element at the end of the current ring
	b.Buffer[(b.Index+b.len)%b.cap] = e
	// Bump up the length
	b.len += 1

	return true
}

// Clear the buffer of all values and allocate a new buffer
func (b *BytesRing) Clear() {
	b.Index = 0
	b.len = 0
	b.Buffer = make([][]byte, b.cap, b.cap)
}

// Len returns the length of the buffer
func (b *BytesRing) Len() int {
	return b.len
}

// Cap returns the capacity of the buffer
func (b *BytesRing) Cap() int {
	return b.cap
}

// SetCap sets the capacity of the buffer and copies all data over
func (b *BytesRing) SetCap(cap int) bool {
	// Cannot set capacity to less than length
	if cap < b.len {
		return false
	}

	// Set the new cap
	b.cap = cap
	// Make the new buffer
	newBuffer := make([][]byte, cap, cap)
	// Copy over the data
	for i := 0; i < b.len; i += 1 {
		newBuffer[i] = b.Buffer[(i+b.Index)%b.cap]
	}
	// Reset the index
	b.Index = 0

	return true
}

// IsEmpty returns true when the length is 0
func (b *BytesRing) IsEmpty() bool {
	return b.len == 0
}

// IsFull returns true when the length is equal to the capacity
func (b *BytesRing) IsFull() bool {
	return b.len == b.cap
}
