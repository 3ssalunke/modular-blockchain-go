package core

import (
	"crypto/elliptic"
	"encoding/gob"
	"io"
)

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobTxEncoder struct {
	w io.Writer
}

func NewGobTxEncoder(w io.Writer) *GobTxEncoder {
	gob.Register(elliptic.P256())
	return &GobTxEncoder{
		w: w,
	}
}

func (e *GobTxEncoder) Encode(tx *Transaction) error {
	enc := gob.NewEncoder(e.w)
	return enc.Encode(tx)
}

type GobTxDecoder struct {
	r io.Reader
}

func NewGobTxDecoder(r io.Reader) *GobTxDecoder {
	gob.Register(elliptic.P256())
	return &GobTxDecoder{
		r: r,
	}
}

func (d *GobTxDecoder) Decode(tx *Transaction) error {
	dec := gob.NewDecoder(d.r)
	return dec.Decode(tx)
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	gob.Register(elliptic.P256())
	return &GobBlockEncoder{
		w: w,
	}
}

func (e *GobBlockEncoder) Encode(b *Block) error {
	enc := gob.NewEncoder(e.w)
	return enc.Encode(b)
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	gob.Register(elliptic.P256())
	return &GobBlockDecoder{
		r: r,
	}
}

func (d *GobBlockDecoder) Decode(b *Block) error {
	dec := gob.NewDecoder(d.r)
	return dec.Decode(b)
}
