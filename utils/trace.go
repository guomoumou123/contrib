package utils

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"sync"
)

type IDGenerator interface {
	NewIDs(ctx context.Context) string
}

type randomIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

var (
	_                IDGenerator = &randomIDGenerator{}
	once             sync.Once
	TraceIDGenerator IDGenerator
)

func init() {
	once.Do(func() {
		TraceIDGenerator = newIDGenerator()
	})
}

func (gen *randomIDGenerator) NewIDs(ctx context.Context) string {
	gen.Lock()
	defer gen.Unlock()
	var tid [16]byte
	_, _ = gen.randSource.Read(tid[:])
	return hex.EncodeToString(tid[:])
}

func newIDGenerator() IDGenerator {
	gen := &randomIDGenerator{}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}
