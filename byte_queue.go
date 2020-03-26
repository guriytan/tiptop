package tiptop

import (
	"encoding/binary"
	"errors"
)

const (
	// Number of bytes used to keep information about entry size
	headerEntrySize = 4
	// Bytes before left margin are not used. Zero index means element does not exist in queue, useful while reading slice from index
	leftMarginIndex = 1
	// Minimum empty blob size in bytes. Empty blob fills space between tail and head in additional memory allocation.
	// It keeps entries indexes unchanged
	minimumEmptyBlobSize = 32 + headerEntrySize
)

var (
	errEmptyQueue       = errors.New("empty queue")
	errInvalidIndex     = errors.New("index must be greater than zero. invalid index")
	errIndexOutOfBounds = errors.New("index out of range")
	errFullQueue        = errors.New("full queue. maximum size limit reached")
)

// ByteQueue is a non-thread safe queue type of fifo based on bytes array.
type ByteQueue struct {
	array           []byte
	capacity        int
	maxCapacity     int
	head            int
	tail            int
	count           int
	rightMargin     int
	headerBuffer    []byte
	initialCapacity int
}

// NewByteQueue initialize new bytes queue.
// Initial capacity is used in bytes array allocation
func NewByteQueue(initialCapacity, maxCapacity int) ByteQueue {
	return ByteQueue{
		array:           make([]byte, initialCapacity),
		capacity:        initialCapacity,
		maxCapacity:     maxCapacity,
		headerBuffer:    make([]byte, headerEntrySize),
		tail:            leftMarginIndex,
		head:            leftMarginIndex,
		rightMargin:     leftMarginIndex,
		initialCapacity: initialCapacity,
	}
}

// Reset removes all entries from queue
func (q *ByteQueue) Reset() {
	// Just reset indexes
	q.tail = leftMarginIndex
	q.head = leftMarginIndex
	q.rightMargin = leftMarginIndex
	q.count = 0
}

// Push copies entry at the end of queue and moves tail pointer. Allocates more space if needed.
// Returns index for pushed data or error if maximum size queue limit is reached.
func (q *ByteQueue) Push(data []byte) (int, error) {
	dataLen := len(data)
	totalLen := dataLen + headerEntrySize

	if q.availableSpaceAfterTail() < totalLen {
		if q.availableSpaceBeforeHead() >= totalLen {
			q.tail = leftMarginIndex
		} else if q.capacity+totalLen >= q.maxCapacity && q.maxCapacity > 0 {
			return -1, errFullQueue
		} else {
			q.allocate(totalLen)
		}
	}

	index := q.tail

	q.push(data, dataLen)

	return index, nil
}

// Pop reads the oldest entry from queue and moves head pointer to the next one
func (q *ByteQueue) Pop() ([]byte, error) {
	data, size, err := q.getWithLength(q.head)
	if err != nil {
		return nil, err
	}

	q.head += headerEntrySize + size
	q.count--

	if q.head == q.rightMargin {
		q.head = leftMarginIndex
		if q.tail == q.rightMargin {
			q.tail = leftMarginIndex
		}
		q.rightMargin = q.tail
	}

	return data, nil
}

// Get reads entry from index
func (q *ByteQueue) Get(index int) ([]byte, error) {
	data, _, err := q.getWithLength(index)
	return data, err
}

// CheckGet checks if an entry can be read from index
func (q *ByteQueue) CheckGet(index int) error {
	return q.peekCheckErr(index)
}

// Peek reads the oldest entry from list without moving head pointer
func (q *ByteQueue) Peek() ([]byte, error) {
	data, _, err := q.getWithLength(q.head)
	return data, err
}

// getWithLength reads entry from index with the data length
func (q *ByteQueue) getWithLength(index int) ([]byte, int, error) {
	err := q.peekCheckErr(index)
	if err != nil {
		return nil, 0, err
	}

	margin := index + headerEntrySize
	blockSize := int(binary.LittleEndian.Uint32(q.array[index:margin]))
	return q.array[margin : margin+blockSize], blockSize, nil
}

// allocate resize the capacity by doubling the original capacity
func (q *ByteQueue) allocate(minimum int) {
	if q.capacity < minimum {
		q.capacity += minimum
	}
	q.capacity = q.capacity * 2
	if q.capacity > q.maxCapacity && q.maxCapacity > 0 {
		q.capacity = q.maxCapacity
	}

	oldArray := q.array
	q.array = make([]byte, q.capacity)

	if leftMarginIndex != q.rightMargin {
		copy(q.array, oldArray[:q.rightMargin])

		if q.tail < q.head {
			// reset the space between the tail and head
			emptyBlobLen := q.head - q.tail - headerEntrySize
			q.push(make([]byte, emptyBlobLen), emptyBlobLen)
			q.head = leftMarginIndex
			q.tail = q.rightMargin
		}
	}
}

func (q *ByteQueue) push(data []byte, len int) {
	binary.LittleEndian.PutUint32(q.headerBuffer, uint32(len))
	q.copy(q.headerBuffer, headerEntrySize)

	q.copy(data, len)

	if q.tail > q.head {
		q.rightMargin = q.tail
	}

	q.count++
}

func (q *ByteQueue) copy(data []byte, len int) {
	q.tail += copy(q.array[q.tail:], data[:len])
}

func (q *ByteQueue) availableSpaceAfterTail() int {
	if q.tail >= q.head {
		return q.capacity - q.tail
	}
	return q.head - q.tail - minimumEmptyBlobSize
}

func (q *ByteQueue) availableSpaceBeforeHead() int {
	if q.tail >= q.head {
		return q.head - leftMarginIndex - minimumEmptyBlobSize
	}
	return q.head - q.tail - minimumEmptyBlobSize
}

// peekCheckErr is identical to peek, but does not actually return any data
func (q *ByteQueue) peekCheckErr(index int) error {

	if q.count == 0 {
		return errEmptyQueue
	}

	if index <= 0 {
		return errInvalidIndex
	}

	if index+headerEntrySize >= len(q.array) {
		return errIndexOutOfBounds
	}
	return nil
}

// Capacity returns number of allocated bytes for queue
func (q *ByteQueue) Capacity() int {
	return q.capacity
}

// Len returns number of entries kept in queue
func (q *ByteQueue) Len() int {
	return q.count
}
